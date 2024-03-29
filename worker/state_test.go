package worker

import (
	"testing"
)

func TestState(t *testing.T) {
	var s *State
	s = &State{Err: "timeout"}
	if s.String() != "timeout" {
		t.Errorf("State invalid, given=%s", s.String())
	}

	s = &State{Err: "timeout", Stderr: "ignore this", Stdout: "also ignore"}
	if s.String() != "timeout: ignore this: also ignore" {
		t.Errorf("State invalid, given=%s", s.String())
	}

	s = &State{Err: "", Stderr: "important so show me", Stdout: "debug info"}
	if s.String() != "important so show me: debug info" {
		t.Errorf("State invalid, given=%s", s.String())
	}

	s = &State{Err: "NotOK", Stderr: "important so show me", Stdout: "debug info"}
	if s.String() != "NotOK: important so show me: debug info" {
		t.Errorf("State invalid, given=%s", s.String())
	}
}

func TestStateHeap(t *testing.T) {
	s := State{Err: "timeout"}
	if s.String() != "timeout" {
		t.Errorf("State invalid, given=%s", s.String())
	}
}
