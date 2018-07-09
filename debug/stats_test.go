package debug

import (
	"os"
	"sync"
	"testing"
)

func setEnv(k, v string) func() {
	prev := os.Getenv(k)
	os.Setenv(k, v)
	return func() {
		os.Setenv(k, prev)
	}
}

func TestNewStats(t *testing.T) {
	reset := setEnv(EnvCFInstanceIndex, "4")
	defer reset()

	stats := NewStats()
	if stats.InstanceIndex != 4 {
		t.Fatalf("TestNewStats: expect %d to be eq 4", stats.InstanceIndex)
	}
}

func TestNewStats_nonNumber(t *testing.T) {
	reset := setEnv(EnvCFInstanceIndex, "ab")
	defer reset()

	stats := NewStats()
	if stats.InstanceIndex != defaultInstanceIndex {
		t.Fatalf("TestNewStats_nonNumber: expect %d to be eq %d", stats.InstanceIndex, defaultInstanceIndex)
	}
}

func TestStatsInc(t *testing.T) {
	s := NewStats()

	loop := 20
	inc := 5

	var wg sync.WaitGroup
	wg.Add(loop)
	for i := 0; i < loop; i++ {
		go func() {
			defer wg.Done()
			for i := 0; i < inc; i++ {
				s.Inc(Consume, 1)
			}
		}()
	}

	wg.Wait()

	expect := loop * inc
	if s.Consume != uint64(expect) {
		t.Fatalf("TestStatsInc: expect %d to be eq %d", s.Consume, expect)
	}
}
