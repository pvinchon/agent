package syncx

import (
	"errors"
	"testing"
)

func TestParallel(t *testing.T) {
	items := []int{1, 2, 3}
	results, errs := Parallel(items, func(n int) (int, error) {
		return n * 2, nil
	})
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	sum := 0
	for _, r := range results {
		sum += r
	}
	if sum != 12 {
		t.Errorf("got sum %d, want 12", sum)
	}
}

func TestParallel_empty(t *testing.T) {
	results, errs := Parallel([]int{}, func(n int) (int, error) {
		return n, nil
	})
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(results) != 0 {
		t.Errorf("got %d results, want 0", len(results))
	}
}

func TestParallel_partialError(t *testing.T) {
	sentinel := errors.New("boom")
	items := []int{1, 2, 3}
	results, errs := Parallel(items, func(n int) (int, error) {
		if n == 2 {
			return 0, sentinel
		}
		return n, nil
	})
	if len(errs) != 1 {
		t.Fatalf("got %d errors, want 1", len(errs))
	}
	if !errors.Is(errs[0], sentinel) {
		t.Errorf("unexpected error: %v", errs[0])
	}
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
}

func TestParallel_allErrors(t *testing.T) {
	sentinel := errors.New("boom")
	items := []int{1, 2}
	results, errs := Parallel(items, func(n int) (int, error) {
		return 0, sentinel
	})
	if len(errs) != 2 {
		t.Fatalf("got %d errors, want 2", len(errs))
	}
	if len(results) != 0 {
		t.Errorf("got %d results, want 0", len(results))
	}
}
