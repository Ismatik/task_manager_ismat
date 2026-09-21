package domain_test

import (
	"fmt"
	"testing"
	"time"

	"nexus/internal/domain"
)

// The C2 benchmark: what it costs to derive a whole board, with the index built
// once and with it rebuilt per card.
//
// # What the two variants are
//
//   - indexed — one domain.Index for the set, then Status and Progress for every
//     node. This is what service.Board() does since S2-01.
//   - perNode — domain.DeriveStatus(set, id) and domain.ComputeProgress(set, id)
//     for every node, each rebuilding the id and children maps from scratch.
//     This is exactly what Board() did before S2-01, and it is kept here rather
//     than described, so the number the commit message quotes can be reproduced.
//
// perNode is O(n²·log n) and really does take minutes at 10 000 nodes; that is
// the point of the measurement. Benchmarks do not run under `go test` without
// -bench, so gate 1 is unaffected:
//
//	go test -run '^$' -bench BenchmarkBoardDerivation -benchtime 1x ./internal/domain/
func BenchmarkBoardDerivation(b *testing.B) {
	for _, size := range []int{100, 1000, 10000} {
		nodes := benchForest(size)

		b.Run(fmt.Sprintf("indexed/%d", size), func(b *testing.B) {
			for b.Loop() {
				ix := domain.NewIndex(nodes)
				for _, n := range nodes {
					if _, err := ix.Status(n.ID); err != nil {
						b.Fatalf("Status(%q): %v", n.ID, err)
					}
					if _, err := ix.Progress(n.ID); err != nil {
						b.Fatalf("Progress(%q): %v", n.ID, err)
					}
				}
			}
		})

		b.Run(fmt.Sprintf("perNode/%d", size), func(b *testing.B) {
			for b.Loop() {
				for _, n := range nodes {
					if _, err := domain.DeriveStatus(nodes, n.ID); err != nil {
						b.Fatalf("DeriveStatus(%q): %v", n.ID, err)
					}
					if _, err := domain.ComputeProgress(nodes, n.ID); err != nil {
						b.Fatalf("ComputeProgress(%q): %v", n.ID, err)
					}
				}
			}
		})
	}
}

// benchForest builds a forest of exactly size nodes shaped like a real board:
// projects of four tasks each, every fifth project nested under the previous one
// so that the derivation recurses instead of reading a flat row of leaves.
func benchForest(size int) []domain.Node {
	const tasksPerProject = 4

	at := time.Date(2026, time.September, 21, 9, 0, 0, 0, time.UTC)
	nodes := make([]domain.Node, 0, size)

	newNode := func(id string, parent *string, typ domain.NodeType, status domain.Status) domain.Node {
		return domain.Node{
			ID:        id,
			ParentID:  parent,
			Type:      typ,
			Title:     id,
			Status:    status,
			DueSource: domain.DueSourceAuto,
			Priority:  domain.Priority4,
			SortOrder: len(nodes),
			CreatedAt: at,
			UpdatedAt: at,
		}
	}

	statuses := domain.Statuses()
	var lastProject *string
	for len(nodes) < size {
		id := fmt.Sprintf("p%05d", len(nodes))
		var parent *string
		if lastProject != nil && len(nodes)%(5*(tasksPerProject+1)) != 0 {
			parent = lastProject
		}
		nodes = append(nodes, newNode(id, parent, domain.NodeTypeProject, domain.StatusBacklog))
		projectID := nodes[len(nodes)-1].ID
		lastProject = &projectID

		for range tasksPerProject {
			if len(nodes) >= size {
				break
			}
			nodes = append(nodes, newNode(
				fmt.Sprintf("t%05d", len(nodes)),
				&projectID,
				domain.NodeTypeTask,
				statuses[len(nodes)%len(statuses)],
			))
		}
	}
	return nodes
}
