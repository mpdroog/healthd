package worker

import (
	"fmt"
	"github.com/mpdroog/healthd/config"
	"strings"
	"sync"
)

type State struct {
	Ok     bool
	Stdout string
	Stderr string
	Err    string
}

func (s *State) String() string {
	var msg string
	if s.Err != "" {
		msg += s.Err
	}
	if len(s.Stderr) > 0 {
		msg += ": " + s.Stderr
	}
	if len(s.Stdout) > 0 {
		msg += ": " + s.Stdout
	}
	if msg[0:1] == ":" {
		msg = msg[2:] // strip prefix
	}
	return strings.Replace(msg, "\n", "", -1)
}

var (
	states     map[string]map[string]*State
	statesLock *sync.RWMutex
)

func initState() error {
	statesLock = new(sync.RWMutex)
	states = nextState()
	if config.Verbose {
		fmt.Printf("worker.States=%+v\n", states)
	}
	states["default"]["healthd"] = &State{Err: "Healthd still starting"}
	return nil
}

func nextState() map[string]map[string]*State {
	s := make(map[string]map[string]*State)
	for dept, _ := range config.Departments {
		d := make(map[string]*State)
		s[dept] = d
	}
	return s
}

// Deep-copy a department
func GetState(dept string) map[string]*State {
	statesLock.RLock()
	defer statesLock.RUnlock()

	n := nextState()
	for k, v := range states[dept] {
		n[dept][k] = v
	}

	return n[dept]
}

// Deep-copy everything
func GetAllStates() map[string]map[string]*State {
	statesLock.RLock()
	defer statesLock.RUnlock()

	n := nextState()
	for dept, vals := range states {
		for k, v := range vals {
			n[dept][k] = v
		}
	}
	return n
}
