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
	Err    error
}

func (s State) String() string {
	var msg string
	if s.Err != nil {
		msg = s.Err.Error()
	} else {
		msg = s.Stderr
		if len(s.Stdout) > 0 {
			msg += s.Stdout
		}
	}
	return strings.Replace(msg, "\n", "", -1)
}

var (
	states     map[string]map[string]State
	statesLock *sync.RWMutex
)

func initState() error {
	statesLock = new(sync.RWMutex)
	states = nextState()
	if config.Verbose {
		fmt.Printf("worker.States=%+v\n", states)
	}
	return nil
}

func nextState() map[string]map[string]State {
	s := make(map[string]map[string]State)
	for dept, _ := range config.Departments {
		d := make(map[string]State)
		s[dept] = d
	}
	return s
}

func GetState(dept string) map[string]State {
	statesLock.RLock()
	defer statesLock.RUnlock()

	return states[dept]
}
func GetAllStates() map[string]map[string]State {
	statesLock.RLock()
	defer statesLock.RUnlock()

	s := states
	return s
}
