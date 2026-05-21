//go:build exercise

package pipeline

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestEndToEnd_SquareThenSum(t *testing.T) {
	ctx := context.Background()
	sum, err := Sum(ctx, Square(ctx, Source(ctx, []int{1, 2, 3, 4})))
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	want := 1 + 4 + 9 + 16
	if sum != want {
		t.Errorf("sum = %d, want %d", sum, want)
	}
}

func TestEmptyInput(t *testing.T) {
	ctx := context.Background()
	sum, err := Sum(ctx, Square(ctx, Source(ctx, nil)))
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	if sum != 0 {
		t.Errorf("sum = %d, want 0", sum)
	}
}

func TestStagesAreComposable(t *testing.T) {
	// Square(Square(Source)) — values come out as n^4.
	ctx := context.Background()
	sum, err := Sum(ctx, Square(ctx, Square(ctx, Source(ctx, []int{1, 2, 3}))))
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	want := 1 + 16 + 81 // 1^4 + 2^4 + 3^4
	if sum != want {
		t.Errorf("sum = %d, want %d", sum, want)
	}
}

func TestSourceClosesItsOutput(t *testing.T) {
	ctx := context.Background()
	out := Source(ctx, []int{10, 20})

	got := make([]int, 0, 2)
	for v := range out {
		got = append(got, v)
	}
	if len(got) != 2 || got[0] != 10 || got[1] != 20 {
		t.Errorf("got %v, want [10 20] (and range must exit cleanly — Source must close its output)", got)
	}
}

// TestSquareClosesItsOutputWhenInputCloses — Square's output should close
// once its input drains, so downstream `range` terminates.
func TestSquareClosesItsOutputWhenInputCloses(t *testing.T) {
	ctx := context.Background()
	in := make(chan int, 3)
	in <- 1
	in <- 2
	in <- 3
	close(in)

	out := Square(ctx, in)
	got := make([]int, 0, 3)
	timeout := time.After(500 * time.Millisecond)
	for {
		select {
		case v, ok := <-out:
			if !ok {
				if len(got) != 3 || got[0] != 1 || got[1] != 4 || got[2] != 9 {
					t.Errorf("got %v, want [1 4 9]", got)
				}
				return
			}
			got = append(got, v)
		case <-timeout:
			t.Fatalf("Square didn't close output after input closed (got %v so far)", got)
		}
	}
}

// TestSumReturnsContextErrorOnCancel — Sum should give up cleanly when ctx
// is cancelled, returning the partial sum and ctx.Err().
func TestSumReturnsContextErrorOnCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	in := make(chan int)
	// Feed two values then sit.
	go func() {
		in <- 5
		in <- 7
		// Don't close. Sum should block here, then ctx cancel below should free it.
	}()

	// Cancel after Sum has had time to consume the 2 values.
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	sum, err := Sum(ctx, in)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("err = nil, want context.Canceled")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, want context.Canceled", err)
	}
	if sum != 12 {
		t.Errorf("partial sum = %d, want 12 (5+7 received before cancel)", sum)
	}
	if elapsed > 500*time.Millisecond {
		t.Errorf("Sum took %v after cancel, expected ~immediate", elapsed)
	}
}
