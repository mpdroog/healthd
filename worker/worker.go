package worker

import (
	"bytes"
	"context"
	"fmt"
	"github.com/mpdroog/healthd/config"
	"io/ioutil"
	"os/exec"
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

// runCmd runs a command with 10sec deadline
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

// Check runs all script.d-files with 3min deadline
func Check() {
	s := nextState()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	for fname, meta := range config.C.Files {
		prefix := fmt.Sprintf("worker(%s)", fname)
		res := runCmd(ctx, fname)
		s[meta.Department][fname] = res

		if config.Verbose {
			fmt.Printf("%s %+v\n", prefix, res)
		}
	}

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
		case <-time.After(5 * time.Minute):
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
}
