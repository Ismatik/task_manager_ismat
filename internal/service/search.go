package service

import (
	"context"
	"strings"

	"nexus/internal/domain"
	"nexus/internal/store"
)

// SearchOptions are the knobs on a search from the application's side.
//
// Tag, type, status and date filters are Stage 3 and deliberately absent: the
// struct is here to be extended, not to be half-filled now with fields no caller
// sets and no test covers.
//
// App.Search takes it, so it crosses the wire and carries the same explicit
// lowerCamelCase tags every other wire type does (S2-02). Without them the
// search input was the only bound shape spelled in PascalCase, and a frontend
// that sent `includeArchived` — the casing everything else uses — would have had
// it silently ignored, which is the worst way for a filter to fail.
type SearchOptions struct {
	// IncludeArchived brings archived nodes back into the results. Off by
	// default, like every other read: archiving is how the user hides something
	// and search is the easiest place to undo that by accident.
	IncludeArchived bool `json:"includeArchived"`

	// Limit caps the number of results. Zero or negative means no cap.
	Limit int `json:"limit"`
}

// SearchService answers the search box, returning the same NodeView cards the
// board and the tree return, so that a result renders exactly like the card it
// is — with its derived status, its progress, its overdue flag and its tags
// already computed.
//
// # The backend is not part of this file
//
// Whether the store searches a full-text index or falls back to something
// simpler is S1-17's business and is invisible here: this service hands the
// user's raw typing to store.SearchRepo and gets []domain.Node back. There is no
// query syntax to build, nothing to escape and no backend to branch on — the
// same test suite passes whichever branch of the S1-04 decision is in force. If
// this file ever needed to know, the store's abstraction would be the thing to
// fix.
type SearchService struct {
	db     Beginner
	nodes  *store.NodeRepo
	tags   *store.TagRepo
	search *store.SearchRepo
	clock  Clock
}

// NewSearchService returns a search service over the repositories and the clock.
func NewSearchService(db Beginner, nodes *store.NodeRepo, tags *store.TagRepo, search *store.SearchRepo, clock Clock) *SearchService {
	return &SearchService{db: db, nodes: nodes, tags: tags, search: search, clock: clock}
}

// Search returns the nodes whose title or description matches query, most
// relevant first, as full cards.
//
// A query with no text in it — empty, or only whitespace — returns an empty
// result and touches no table at all. Returning everything for an empty search
// box is how a search feature becomes a way to accidentally list the whole
// database, and it is also the state the box is in every time it is opened.
//
// The results and the tree they are derived over are read in one transaction, so
// a card's status cannot be derived from a set that changed halfway through.
func (s *SearchService) Search(ctx context.Context, query string, opts SearchOptions) ([]NodeView, error) {
	if strings.TrimSpace(query) == "" {
		return []NodeView{}, nil
	}

	out := []NodeView{}
	err := runInTx(ctx, s.db, func(exec store.Executor) error {
		hits, err := s.search.WithExecutor(exec).Search(ctx, query, store.SearchOptions{
			IncludeArchived: opts.IncludeArchived,
			Limit:           opts.Limit,
		})
		if err != nil {
			return err
		}
		if len(hits) == 0 {
			return nil
		}

		// The derivations need the tree the hits live in, not just the hits: a
		// parent's status comes from its children, and a child that did not
		// match the query is still one of them. The set is loaded under the
		// same archived filter as the search itself, so a card in a result
		// derives exactly what it derives on the board.
		set, err := s.nodes.WithExecutor(exec).ListAll(ctx, opts.IncludeArchived)
		if err != nil {
			return err
		}

		snap, err := loadSnapshot(ctx, exec, set, s.clock)
		if err != nil {
			return err
		}

		out = make([]NodeView, 0, len(hits))
		for _, n := range hits {
			v, err := snap.view(n)
			if err != nil {
				return err
			}
			out = append(out, v)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Rebuild reindexes every node.
//
// It is a recovery path, not routine maintenance: the index is kept current by
// the triggers migration 0003 installs, and this is what a user needs after
// restoring a database file or after anything that renumbers rowids. It is
// exposed here rather than left in the store so that the one call the settings
// screen will eventually make is a service call like every other.
func (s *SearchService) Rebuild(ctx context.Context) error {
	return runInTx(ctx, s.db, func(exec store.Executor) error {
		return s.search.WithExecutor(exec).Rebuild(ctx)
	})
}

// searchable is a compile-time reminder of what this file is allowed to know
// about the backend: a query string in, domain nodes out.
var _ interface {
	Search(ctx context.Context, query string, opts store.SearchOptions) ([]domain.Node, error)
} = (*store.SearchRepo)(nil)
