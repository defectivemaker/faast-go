package permute

import (
	"reflect"
	"testing"
	"time"
)

func TestNewPermutationIterator(t *testing.T) {
	lists := [][]string{{"a", "b"}, {"1", "2"}}
	pi := NewPermutationIterator(lists)

	if pi == nil {
		t.Fatal("NewPermutationIterator returned nil")
	}

	if !reflect.DeepEqual(pi.lists, lists) {
		t.Errorf("Expected lists %v, got %v", lists, pi.lists)
	}

	if len(pi.indices) != len(lists) {
		t.Errorf("Expected indices length %d, got %d", len(lists), len(pi.indices))
	}

	for _, index := range pi.indices {
		if index != 0 {
			t.Errorf("Expected all indices to be 0, got %d", index)
		}
	}

	if pi.finished {
		t.Error("Expected finished to be false")
	}
}

func TestPermutationIterator_Next(t *testing.T) {
	lists := [][]string{{"a", "b"}, {"1", "2"}}
	pi := NewPermutationIterator(lists)

	expected := [][]string{
		{"a", "1"},
		{"a", "2"},
		{"b", "1"},
		{"b", "2"},
	}

	for i, exp := range expected {
		result, ok := pi.Next()
		if !ok {
			t.Fatalf("Expected ok to be true for iteration %d", i)
		}
		if !reflect.DeepEqual(result, exp) {
			t.Errorf("Iteration %d: expected %v, got %v", i, exp, result)
		}
	}

	// Check that we're finished
	result, ok := pi.Next()
	if ok || result != nil {
		t.Errorf("Expected (nil, false), got (%v, %v)", result, ok)
	}
}

func TestShardLists(t *testing.T) {
	tests := []struct {
		name        string
		lists       [][]string
		n           int
		numOfShards int
		want        [][][]string
	}{
		{
			name:        "Basic sharding",
			lists:       [][]string{{"a", "b"}, {"1", "2", "3", "4"}},
			n:           1,
			numOfShards: 2,
			want: [][][]string{
				{{"a", "b"}, {"1", "2"}},
				{{"a", "b"}, {"3", "4"}},
			},
		},
		{
			name:        "Empty input",
			lists:       [][]string{},
			n:           0,
			numOfShards: 2,
			want:        [][][]string{},
		},
		{
			name:        "Invalid n",
			lists:       [][]string{{"a", "b"}, {"1", "2"}},
			n:           2,
			numOfShards: 2,
			want:        [][][]string{},
		},
		{
			name:        "Invalid numOfShards",
			lists:       [][]string{{"a", "b"}, {"1", "2"}},
			n:           0,
			numOfShards: 0,
			want:        [][][]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShardLists(tt.lists, tt.n, tt.numOfShards)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ShardLists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculateTotalPermutations(t *testing.T) {
	tests := []struct {
		name         string
		shardedLists [][][]string
		want         int
	}{
		{
			name: "Basic calculation",
			shardedLists: [][][]string{
				{{"a", "b"}, {"1", "2"}},
				{{"a", "b"}, {"3", "4"}},
			},
			want: 8,
		},
		{
			name:         "Empty input",
			shardedLists: [][][]string{},
			want:         0,
		},
		{
			name: "Single empty list",
			shardedLists: [][][]string{
				{{}},
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateTotalPermutations(tt.shardedLists)
			if got != tt.want {
				t.Errorf("CalculateTotalPermutations() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIteratePermutations(t *testing.T) {
	lists := [][]string{{"a", "b"}, {"1", "2"}}
	pi := NewPermutationIterator(lists)

	results := make(chan []string)
	done := make(chan struct{})

	go func() {
		IteratePermutations(pi, results)
		close(results)
	}()

	expected := [][]string{
		{"a", "1"},
		{"a", "2"},
		{"b", "1"},
		{"b", "2"},
	}

	go func() {
		for i, exp := range expected {
			result, ok := <-results
			if !ok {
				t.Errorf("Channel closed before receiving all expected results. Got %d out of %d", i, len(expected))
				break
			}
			if !reflect.DeepEqual(result, exp) {
				t.Errorf("IteratePermutations() yielded %v, want %v", result, exp)
			}
		}

		// Check if there are any extra results
		extra, ok := <-results
		if ok {
			t.Errorf("IteratePermutations() yielded extra result: %v", extra)
			// Drain the channel
			for range results {
			}
		}

		close(done)
	}()

	// Wait for all checks to complete
	select {
	case <-done:
		// Test completed successfully
	case <-time.After(5 * time.Second):
		t.Error("Test timed out")
	}
}
