package utils

import (
	"container/heap"
	"testing"

	"github.com/qaynaq/qaynaq/internal/persistence"
)

func makeWorker(id string, count int) persistence.Worker {
	return persistence.Worker{ID: id, RunningFlowCount: count}
}

func popIDs(t *testing.T, h *WorkerHeap, n int) []string {
	t.Helper()
	ids := make([]string, 0, n)
	for range n {
		w := heap.Pop(h).(persistence.Worker)
		ids = append(ids, w.ID)
	}
	return ids
}

func TestWorkerHeap_PopReturnsLeastLoaded(t *testing.T) {
	h := &WorkerHeap{}
	heap.Init(h)
	heap.Push(h, makeWorker("a", 3))
	heap.Push(h, makeWorker("b", 1))
	heap.Push(h, makeWorker("c", 2))

	got := popIDs(t, h, 3)
	want := []string{"b", "c", "a"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("pop %d: want %s, got %s (full order %v)", i, want[i], got[i], got)
		}
	}
}

func TestWorkerHeap_EmptyAfterDraining(t *testing.T) {
	h := &WorkerHeap{}
	heap.Push(h, makeWorker("a", 0))
	heap.Pop(h)
	if h.Len() != 0 {
		t.Fatalf("expected empty heap, got len %d", h.Len())
	}
}

func TestWorkerHeap_AssignmentLoopBalancesLoad(t *testing.T) {
	h := &WorkerHeap{}
	heap.Init(h)
	for _, id := range []string{"a", "b", "c", "d"} {
		heap.Push(h, makeWorker(id, 0))
	}

	const flows = 9
	assignments := map[string]int{}
	for range flows {
		w := heap.Pop(h).(persistence.Worker)
		w.RunningFlowCount++
		assignments[w.ID]++
		heap.Push(h, w)
	}

	lo, hi := flows, 0
	for _, c := range assignments {
		lo = min(lo, c)
		hi = max(hi, c)
	}
	if hi-lo > 1 {
		t.Fatalf("load not balanced: %v (spread %d)", assignments, hi-lo)
	}
}

func TestWorkerHeap_NextPickIsTheJustIncrementedWorker(t *testing.T) {
	h := &WorkerHeap{}
	heap.Init(h)
	heap.Push(h, makeWorker("low", 0))
	heap.Push(h, makeWorker("hi1", 5))
	heap.Push(h, makeWorker("hi2", 5))

	w := heap.Pop(h).(persistence.Worker)
	if w.ID != "low" {
		t.Fatalf("first pop: want low, got %s", w.ID)
	}
	w.RunningFlowCount++
	heap.Push(h, w)

	next := heap.Pop(h).(persistence.Worker)
	if next.ID != "low" {
		t.Fatalf("second pop: want low (count 1), got %s (count %d)", next.ID, next.RunningFlowCount)
	}
}

func TestWorkerHeap_EqualCountsAreStable(t *testing.T) {
	h := &WorkerHeap{}
	heap.Init(h)
	heap.Push(h, makeWorker("a", 2))
	heap.Push(h, makeWorker("b", 2))
	heap.Push(h, makeWorker("c", 2))

	seen := map[string]bool{}
	for range 3 {
		w := heap.Pop(h).(persistence.Worker)
		if seen[w.ID] {
			t.Fatalf("worker %s popped twice", w.ID)
		}
		seen[w.ID] = true
	}
	if len(seen) != 3 {
		t.Fatalf("expected 3 distinct workers, got %d", len(seen))
	}
}
