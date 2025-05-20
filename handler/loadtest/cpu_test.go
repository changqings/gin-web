package loadtest

import "testing"

func TestBurnCpu(t *testing.T) {
	l := NewLoadTestCpu()
	l.DurationSeconds = 25

	burnCpu(l.DurationSeconds, l.workers)
	t.Log("burn cpu success")
}
