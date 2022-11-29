package worker

import (
	"bytes"
	"context"
	"fmt"
	"github.com/mpdroog/healthd/config"
	"io/ioutil"
	"os/exec"
	"strings"
	"sync"
	"time"
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

func Init() error {
	statesLock = new(sync.RWMutex)
	states = make(map[string]map[string]State)
	for dept, _ := range config.Departments {
		d := make(map[string]State)
		states[dept] = d
	}
	if config.Verbose {
		fmt.Printf("worker.States=%+v\n", states)
	}

	go loop()
	return nil
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

func runCmd(ctxGroup context.Context, fname string) State {
	ctx, cancel := context.WithTimeout(ctxGroup, 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, fname, "")

	re, e := cmd.StderrPipe()
	if e != nil {
		return State{Err: e}
	}

	ro, e := cmd.StdoutPipe()
	if e != nil {
		return State{Err: e}
	}

	if e := cmd.Start(); e != nil {
		return State{Err: e}
	}

	stdErr, e := ioutil.ReadAll(re)
	if e != nil {
		return State{Err: e}
	}
	stdOut, e := ioutil.ReadAll(ro)
	if e != nil {
		return State{Err: e}
	}

	// Ignore Wait-output
	cmd.Wait()
	ok := bytes.HasPrefix(stdOut, []byte("OK"))
	return State{
		Ok:     ok,
		Stdout: string(stdOut),
		Stderr: string(stdErr),
	}
}

func Check() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	for fname, meta := range config.C.Files {
		prefix := fmt.Sprintf("worker(%s)", fname)
		res := runCmd(ctx, fname)

		statesLock.Lock()
		states[meta.Department][fname] = res
		statesLock.Unlock()

		if config.Verbose {
			fmt.Printf("%s %+v\n", prefix, res)
		}
	}
}

// Run every 5mins and remember state
func loop() {
	Check()

	for {
		if config.Verbose {
			fmt.Println("worker sleep 5min")
		}

		select {
		case <-time.After(5 * time.Minute):
			if config.Verbose {
				fmt.Println("5min passed")
			}
			if e := config.ReloadConf(); e != nil {
				fmt.Printf("WARN: config.ReloadConf e=%s\n", e.Error())
			}
			Check()
		}
	}
}
