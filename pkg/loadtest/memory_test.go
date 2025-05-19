package loadtest

import "testing"

func TestLoadMemory(t *testing.T) {

	l := NewLoadtestMemory()
	l.Duration = 25

	burnMemory(l.Duration, l.SizeMB)
	t.Log("TestLoadMemory complete")
}
