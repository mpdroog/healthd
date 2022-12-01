package worker

import (
	"bytes"
	"context"
	"fmt"
	"github.com/mpdroog/healthd/config"
	"io/ioutil"
	"sync"
	"time"
)

// Init prepares memory and spawns the go-routine
func Init() error {
	if e := initState(); e != nil {
		return e
	}
	go loop()
	return nil
}

// runCmd runs a command with 1min deadline
func runCmd(ctxGroup context.Context, fname string) (s *State) {
	s = new(State)
	ctx, cancel := context.WithTimeout(ctxGroup, 1*time.Minute)
	defer cancel()
	cmd := NewCommand(ctx, fname)

	re, e := cmd.StderrPipe()
	if e != nil {
		s.Err = e.Error()
		return
	}

	ro, e := cmd.StdoutPipe()
	if e != nil {
		s.Err = e.Error()
		return
	}

	if e := cmd.Start(); e != nil {
		s.Err = e.Error()
		return
	}

	stdErr, e := ioutil.ReadAll(re)
	if e != nil {
		s.Err = e.Error()
	}
	stdOut, e := ioutil.ReadAll(ro)
	if e != nil {
		s.Err = e.Error()
	}

	if e := cmd.Wait(); e != nil {
		s.Err = e.Error()
	}

	s.Stdout = string(bytes.TrimSpace(stdOut))
	s.Stderr = string(bytes.TrimSpace(stdErr))
	s.Ok = bytes.HasPrefix(stdOut, []byte("OK"))

	if !s.Ok && s.Err == "" {
		s.Err = "Stdout not OK"
	}
	return s
}

// Check runs all script.d-files with 3min deadline
func Check() {
	s := nextState()
	sLock := new(sync.Mutex)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	if len(config.C.Files) == 0 {
		s["default"]["healthd"] = &State{Err: "Misconfig: No scripts to run"}
	}

	wg := new(sync.WaitGroup)
	// TODO: Cleaner go-routines?
	for fname, meta := range config.C.Files {
		wg.Add(1)
		go func(fname string, meta config.File) {
			prefix := fmt.Sprintf("worker(%s)", fname)
			res := runCmd(ctx, fname)

			sLock.Lock()
			if _, already := s[meta.Department][fname]; already {
				s["default"]["healthd"] = &State{Err: "Misconfig: Double-run " + fname}
			}
			s[meta.Department][fname] = res
			if config.Verbose {
				fmt.Printf("%s: %+v\n", prefix, res)
			}
			sLock.Unlock()
			wg.Done()
		}(fname, meta)
	}
	wg.Wait()

	if config.Verbose {
		fmt.Printf("Result=%+v\n", s)
	}
	statesLock.Lock()
	states = s
	statesLock.Unlock()
}

// Run every 5mins and remember state
func loop() {
	if config.Verbose {
		fmt.Println("worker.Check(initial)")
	}
	Check()

	for {
		if config.Verbose {
			fmt.Println("worker sleep 5min")
		}

		select {
			case <-config.RefreshChan:
				fmt.Println("Refresh signal")
			case <-time.After(5 * time.Minute):
		}
		if config.Verbose {
			fmt.Println("worker.Reload")
		}
		if e := config.ReloadConf(); e != nil {
			fmt.Printf("WARN: config.ReloadConf e=%s\n", e.Error())
		}

		if config.Verbose {
			fmt.Println("worker.Check(after 5min sleep)")
		}
		Check()
	}
}
