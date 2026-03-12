package sync

import "sync"

// Parallel runs fn on each item concurrently and returns the collected results
// and any errors.
func Parallel[T, R any](items []T, fn func(T) (R, error)) ([]R, []error) {
	type result struct {
		item R
		err  error
	}

	results := make([]result, len(items))
	var wg sync.WaitGroup
	wg.Add(len(items))

	for i, item := range items {
		go func() {
			defer wg.Done()
			r, err := fn(item)
			results[i] = result{item: r, err: err}
		}()
	}

	wg.Wait()

	var all []R
	var errs []error
	for _, res := range results {
		if res.err != nil {
			errs = append(errs, res.err)
		} else {
			all = append(all, res.item)
		}
	}
	return all, errs
}
