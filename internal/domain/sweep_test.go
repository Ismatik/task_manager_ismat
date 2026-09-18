package domain_test

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"testing"

	"nexus/internal/domain"
)

// A randomised sweep over generated forests, checking the invariants of the
// no-column rule against a REFERENCE WALK that shares no code with the
// implementation it is checking (S1-09).
//
// # Why this file exists
//
// Five review failures in a row were the same defect: a type rule spelled in two
// places, one copy left behind. Every example-based test in this package asserts
// a shape somebody thought of; this one asserts a property over shapes nobody
// thought of, which is the only kind of test that catches the copy that was
// missed rather than the copy that was fixed. The reviewer's sweep found the
// S1-09 defect and was not committed anywhere, so it guarded nothing afterwards.
// This is a bounded version of it that runs inside `make check`.
//
// # Bounded on purpose
//
// forestCount forests of up to maxNodes nodes runs in well under a second, which
// is what keeps it in the standard gate rather than in a nightly job nobody
// runs. Set NEXUS_SWEEP_FORESTS to a larger number to run the wide version by
// hand; the generator is seeded per forest, so forest number k is the same tree
// whatever the total is, and a failure names the seed that produced it.
const (
	forestCount = 3000
	maxNodes    = 12
)

func sweepForests(t *testing.T) int {
	t.Helper()

	raw := os.Getenv("NEXUS_SWEEP_FORESTS")
	if raw == "" {
		return forestCount
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		t.Fatalf("NEXUS_SWEEP_FORESTS = %q: want a positive integer", raw)
	}
	return n
}

// forest is one generated node set and the way it was built.
type forest struct {
	nodes []domain.Node
	seed  int64
}

// generate builds one pseudo-random forest.
//
// gated decides what a parent may be: when true every parent is proposed to
// domain.ValidateMove and rejected proposals fall back to the root, which is
// exactly the door CreateNode and MoveNode go through — so a gated forest is one
// the application could really produce. When false any node may parent any
// other, including the shapes the rules forbid, which is how corrupt or migrated
// data is simulated.
func generate(seed int64, gated bool) forest {
	rng := rand.New(rand.NewSource(seed))
	types := domain.NodeTypes()

	n := 1 + rng.Intn(maxNodes)
	nodes := make([]domain.Node, 0, n)

	for i := 0; i < n; i++ {
		node := domain.Node{
			ID:        fmt.Sprintf("n%d", i),
			Type:      types[rng.Intn(len(types))],
			Title:     fmt.Sprintf("n%d", i),
			Status:    domain.StatusBacklog,
			DueSource: domain.DueSourceManual,
			Priority:  domain.Priority4,
			CreatedAt: fixedNow,
			UpdatedAt: fixedNow,
		}
		if node.Type == domain.NodeTypeHabit {
			node.Recurrence = ptr("FREQ=DAILY")
		}

		// A root two times in five, otherwise somebody already placed.
		if i > 0 && rng.Intn(5) > 1 {
			candidate := nodes[rng.Intn(len(nodes))].ID
			switch {
			case !gated:
				node.ParentID = ptr(candidate)
			case domain.ValidateMove(append(append([]domain.Node{}, nodes...), node), node.ID, &candidate) == nil:
				node.ParentID = ptr(candidate)
			}
		}
		nodes = append(nodes, node)
	}

	assignStatuses(rng, nodes, gated)
	return forest{nodes: nodes, seed: seed}
}

// assignStatuses writes a plausible stored status onto every node: the rules the
// write paths enforce (domain.CheckStatus), applied here with the generator's
// own child map so that a forest looks like one the application stored.
//
// An UNGATED forest is corrupt by construction, so a type with no column gets a
// real column written onto it now and then — a row CheckStatus refuses and an
// older build could still have left behind. That is what makes the ungated sweep
// a check on the self cut in deriveStatus rather than only on Node.IsLeaf: a
// habit stored at today must render backlog, with children or without.
func assignStatuses(rng *rand.Rand, nodes []domain.Node, gated bool) {
	hasColumnChild := make(map[string]bool, len(nodes))
	for _, n := range nodes {
		if n.ParentID != nil && n.Type.HasColumn() {
			hasColumnChild[*n.ParentID] = true
		}
	}

	for i := range nodes {
		n := &nodes[i]
		if !n.Type.HasColumn() {
			// The inert value its NOT NULL column needs; it means "no column".
			n.Status = domain.StatusBacklog
			if !gated && rng.Intn(2) == 0 {
				n.Status = domain.Statuses()[rng.Intn(len(domain.Statuses()))]
			}
			continue
		}
		// doing is stored only where a timer may run: a leaf that is not a
		// project (D2, D9). Everything else draws from the other four.
		choices := []domain.Status{
			domain.StatusBacklog, domain.StatusWeek, domain.StatusToday, domain.StatusDone,
		}
		if n.Type != domain.NodeTypeProject && !hasColumnChild[n.ID] {
			choices = append(choices, domain.StatusDoing)
		}
		n.Status = choices[rng.Intn(len(choices))]
	}
}

// reference is an independent reading of a node set: its own child map, its own
// leaf rule, its own derivation. Nothing in here calls the package under test,
// which is the point — an invariant checked with the implementation's own walk
// only proves the walk is self-consistent.
type reference struct {
	byID map[string]domain.Node
	kids map[string][]string
}

func newReference(nodes []domain.Node) *reference {
	r := &reference{
		byID: make(map[string]domain.Node, len(nodes)),
		kids: make(map[string][]string, len(nodes)),
	}
	for _, n := range nodes {
		r.byID[n.ID] = n
		key := ""
		if n.ParentID != nil {
			key = *n.ParentID
		}
		r.kids[key] = append(r.kids[key], n.ID)
	}
	return r
}

// noColumn is the reference's own spelling of the one rule under test. It is
// written out as a list of types deliberately: if NodeType.HasColumn itself ever
// drifts, this sweep must notice rather than drift with it.
func (r *reference) noColumn(t domain.NodeType) bool {
	return t == domain.NodeTypeNote || t == domain.NodeTypeHabit
}

// isLeaf: nothing under it that has a column, or no column of its own.
func (r *reference) isLeaf(id string) bool {
	if r.noColumn(r.byID[id].Type) {
		return true
	}
	for _, c := range r.kids[id] {
		if !r.noColumn(r.byID[c].Type) {
			return false
		}
	}
	return true
}

// rank orders the statuses from least to most advanced, written out rather than
// read from domain.Statuses().
var rank = map[domain.Status]int{
	domain.StatusBacklog: 0,
	domain.StatusWeek:    1,
	domain.StatusToday:   2,
	domain.StatusDoing:   3,
	domain.StatusDone:    4,
}

// derive is the reference reading of D2 + D10 + S1-09: a node with no column
// reads backlog; a leaf reads its stored status; a parent reads the
// least-advanced status among its column-bearing children, done when all of
// them are.
func (r *reference) derive(id string) domain.Status {
	n := r.byID[id]
	if r.noColumn(n.Type) {
		return domain.StatusBacklog
	}
	if r.isLeaf(id) {
		return n.Status
	}

	least := domain.Status("")
	allDone := true
	for _, c := range r.kids[id] {
		if r.noColumn(r.byID[c].Type) {
			continue
		}
		s := r.derive(c)
		if s == domain.StatusDone {
			continue
		}
		allDone = false
		if least == "" || rank[s] < rank[least] {
			least = s
		}
	}
	if allDone {
		return domain.StatusDone
	}
	return least
}

// unfinishedLeafBelow returns the first LEAF below id that has a column and is
// not done, or "".
//
// Only leaves are asked because only a leaf's stored status is the truth: a
// parent's is never written and means nothing. The walk does NOT stop at a
// no-column node — the board shows every card whose type has a column wherever
// it sits in the tree, so a task hidden under a habit is still on screen and
// still unfinished.
func (r *reference) unfinishedLeafBelow(id string) string {
	for _, c := range r.kids[id] {
		n := r.byID[c]
		if !r.noColumn(n.Type) && r.isLeaf(c) && n.Status != domain.StatusDone {
			return c
		}
		if found := r.unfinishedLeafBelow(c); found != "" {
			return found
		}
	}
	return ""
}

// violations counts a sweep's failures by the invariant that failed, so that a
// red run says WHICH property broke and how often rather than only that
// something did. The first few are printed with the seed that produced them; a
// seed plus this file reproduces the forest exactly.
type violations struct {
	t     *testing.T
	count map[string]int
	order []string
	shown int
}

func newViolations(t *testing.T) *violations {
	return &violations{t: t, count: map[string]int{}}
}

func (v *violations) add(invariant string, seed int64, format string, a ...any) {
	if _, seen := v.count[invariant]; !seen {
		v.order = append(v.order, invariant)
	}
	v.count[invariant]++

	if v.shown < 10 {
		v.shown++
		v.t.Errorf("%s violated, seed %d: %s", invariant, seed, fmt.Sprintf(format, a...))
	}
}

func (v *violations) total() int {
	total := 0
	for _, n := range v.count {
		total += n
	}
	return total
}

func (v *violations) report() string {
	out := ""
	for _, name := range v.order {
		if out != "" {
			out += ", "
		}
		out += fmt.Sprintf("%s x%d", name, v.count[name])
	}
	return out
}

// The invariants, named so that the counts above mean something.
const (
	invShape    = "I1 no node sits under a type with no column"
	invNoColumn = "I2 a type with no column derives backlog and nothing else"
	invWalk     = "I3 derivation matches the reference walk"
	invAgree    = "I4 the derived column and the progress bar agree"
	invBuried   = "I5 nothing reads done at 100% over unfinished work"
	invLeaf     = "I6 a type with no column is a leaf"
	invError    = "I0 the walks return no error"
)

// TestSweepLegalForests: forests built the only way the application can build
// them — every parent approved by domain.ValidateMove.
//
// The invariants are the reviewer's. Before S1-09 the second one failed roughly
// 32,300 times and the third roughly 800 times in 200,000 forests, every one of
// them through the same door: a parent whose type has no Kanban column.
func TestSweepLegalForests(t *testing.T) {
	count := sweepForests(t)
	bad := newViolations(t)
	// Counted so that a sweep which happens to check nothing fails loudly
	// instead of reporting a green zero.
	nested, noColumnNodes, doneAndFull := 0, 0, 0

	for i := 0; i < count; i++ {
		f := generate(int64(i), true)
		ref := newReference(f.nodes)

		for _, n := range f.nodes {
			if n.ParentID != nil {
				nested++
			}
			if ref.noColumn(n.Type) {
				noColumnNodes++
			}

			// The shape itself: a gated forest cannot contain a node parked
			// under a type with no column.
			if n.ParentID != nil && ref.noColumn(ref.byID[*n.ParentID].Type) {
				bad.add(invShape, f.seed, "%q sits under the %s %q, which has no column",
					n.ID, ref.byID[*n.ParentID].Type, *n.ParentID)
			}

			status, err := domain.DeriveStatus(f.nodes, n.ID)
			if err != nil {
				bad.add(invError, f.seed, "DeriveStatus(%q) = %v", n.ID, err)
				continue
			}
			progress, err := domain.ComputeProgress(f.nodes, n.ID)
			if err != nil {
				bad.add(invError, f.seed, "ComputeProgress(%q) = %v", n.ID, err)
				continue
			}

			if ref.noColumn(n.Type) && status != domain.StatusBacklog {
				bad.add(invNoColumn, f.seed, "the %s %q derives %q, want backlog",
					n.Type, n.ID, status)
			}
			if want := ref.derive(n.ID); status != want {
				bad.add(invWalk, f.seed, "DeriveStatus(%q) = %q, the reference walk says %q",
					n.ID, status, want)
			}
			if progress.Defined() && (status == domain.StatusDone) != (progress.Percent() == 100) {
				bad.add(invAgree, f.seed, "%q renders %q with a bar at %d%%",
					n.ID, status, progress.Percent())
			}
			if status == domain.StatusDone && progress.Defined() && progress.Percent() == 100 {
				doneAndFull++
				if buried := ref.unfinishedLeafBelow(n.ID); buried != "" {
					bad.add(invBuried, f.seed,
						"%q reads done at 100%% while the leaf %q (%s, %q) is unfinished below it",
						n.ID, buried, ref.byID[buried].Type, ref.byID[buried].Status)
				}
			}
		}
	}

	if nested == 0 || noColumnNodes == 0 || doneAndFull == 0 {
		t.Fatalf("the sweep checked nothing worth checking: %d nested nodes, %d with no column, "+
			"%d reading done at 100%%", nested, noColumnNodes, doneAndFull)
	}
	if total := bad.total(); total > 0 {
		t.Fatalf("%d invariant violations over %d gated forests: %s", total, count, bad.report())
	}
	t.Logf("0 violations over %d gated forests (%d nested nodes, %d with no column, "+
		"%d reading done at 100%%)", count, nested, noColumnNodes, doneAndFull)
}

// TestSweepCorruptForests: forests in which anything may parent anything,
// including the shapes ValidateMove now refuses.
//
// This is the half that guards the SELF cut. No door can produce these node sets
// any more — which is precisely why the derivation must be right about them
// anyway, since a database written by an older build can hold one. Only the
// invariants that survive corrupt data are asserted: a hidden task really is
// unfinished work below a done project here, and refusing to create it is the
// only fix for that, which is what the gated sweep above measures.
func TestSweepCorruptForests(t *testing.T) {
	count := sweepForests(t)
	bad := newViolations(t)
	withNoColumnParent := 0

	for i := 0; i < count; i++ {
		f := generate(int64(i), false)
		ref := newReference(f.nodes)

		for _, n := range f.nodes {
			if n.ParentID != nil && ref.noColumn(ref.byID[*n.ParentID].Type) {
				withNoColumnParent++
			}

			status, err := domain.DeriveStatus(f.nodes, n.ID)
			if err != nil {
				bad.add(invError, f.seed, "DeriveStatus(%q) = %v", n.ID, err)
				continue
			}
			progress, err := domain.ComputeProgress(f.nodes, n.ID)
			if err != nil {
				bad.add(invError, f.seed, "ComputeProgress(%q) = %v", n.ID, err)
				continue
			}

			if ref.noColumn(n.Type) && status != domain.StatusBacklog {
				bad.add(invNoColumn, f.seed, "the %s %q derives %q, want backlog",
					n.Type, n.ID, status)
			}
			if want := ref.derive(n.ID); status != want {
				bad.add(invWalk, f.seed, "DeriveStatus(%q) = %q, the reference walk says %q",
					n.ID, status, want)
			}
			if progress.Defined() && (status == domain.StatusDone) != (progress.Percent() == 100) {
				bad.add(invAgree, f.seed, "%q renders %q with a bar at %d%%",
					n.ID, status, progress.Percent())
			}
			if ref.noColumn(n.Type) && !n.IsLeaf(domain.Children(f.nodes, n.ID)) {
				bad.add(invLeaf, f.seed, "the %s %q is not a leaf, but nothing may be under it",
					n.Type, n.ID)
			}
		}
	}

	if withNoColumnParent == 0 {
		t.Fatalf("no forest held a node under a no-column parent: the sweep checked nothing")
	}
	if total := bad.total(); total > 0 {
		t.Fatalf("%d invariant violations over %d ungated forests: %s", total, count, bad.report())
	}
	t.Logf("0 violations over %d ungated forests (%d nodes sat under a no-column parent)",
		count, withNoColumnParent)
}
