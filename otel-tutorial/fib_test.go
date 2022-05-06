package main

import (
	"testing"
)

func TestFib(t *testing.T) {
	tables := []struct {
		x uint
		n uint64
	}{
		{0, 0},
		{1, 1},
		{2, 1},
		{3, 2},
		{10, 55},
		{93, 12200160415121876738},
	}
	for _, table := range tables {
		val, err := Fibonacci(table.x)
		if err != nil {
			t.Errorf("Fibonacci(%d) returned error!", table.x)
		} else if val != table.n {
			t.Errorf("Fibonacci(%d) is %d. Received %d instead.", table.x, table.n, val)
		}
	}
	_, err := Fibonacci(94)
	if err == nil {
		t.Errorf("Expected Fibonacci(94) to return error.")
	}
}
