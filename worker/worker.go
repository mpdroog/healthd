package worker

import (
	"bytes"
	"fmt"
	"github.com/mpdroog/healthd/config"
	"io/ioutil"
	"os/exec"
	"strings"
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

var States map[string]map[string]State

func Init() error {
	States = make(map[string]map[string]State)
	for dept, _ := range config.Departments {
		d := make(map[string]State)
		States[dept] = d
	}
	if config.Verbose {
		fmt.Printf("worker.States=%+v\n", States)
	}

	go loop()
	return nil
}

func runCmd(fname string) State {
	cmd := exec.Command(fname, "")

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
	for fname, meta := range config.C.Files {
		prefix := fmt.Sprintf("worker(%s)", fname)
		States[meta.Department][fname] = runCmd(fname)
		if config.Verbose {
			fmt.Printf("%s %+v\n", prefix, States[fname])
		}
	}
}

// Run every 5mins and remember state
// for Zenoss
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
			Check()
		}
	}
}
