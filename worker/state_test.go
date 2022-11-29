package worker

import (
	"testing"
	"fmt"
)

func TestState(t *testing.T) {
	var s *State
	s = &State{Err: fmt.Errorf("timeout")}
	if s.String() != "timeout" {
		t.Errorf("State invalid, given=%s", s.String())
	}

	s = &State{Err: fmt.Errorf("timeout"), Stderr: "ignore this", Stdout: "also ignore"}
	if s.String() != "timeout" {
		t.Errorf("State invalid, given=%s", s.String())
	}

	s = &State{Err: nil, Stderr: "important so show me", Stdout: "debug info"}
	if s.String() != "important so show me, debug info" {
		t.Errorf("State invalid, given=%s", s.String())
	}
}

func TestStateHeap(t *testing.T) {
	s := State{Err: fmt.Errorf("timeout")}
	if s.String() != "timeout" {
		t.Errorf("State invalid, given=%s", s.String())
	}
}