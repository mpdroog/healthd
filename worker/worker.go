package worker

import (
	"github.com/mpdroog/healthd/config"
	"time"
	"os/exec"
	"fmt"
	"io/ioutil"
	"bytes"
)

type State struct {
	prefix string

	Ok bool
	Stdout string
	Stderr string
	Err error
}

func (s State) String() string {
	msg := s.Stderr
	if len(s.Stdout) > 0 {
		msg += s.Stdout
	}
	if s.Err != nil {
		msg += s.Err.Error()
	}
	return fmt.Sprintf("%s: %s", s.prefix, msg)
}

var States map[string]State

func Init() error {
	States = make(map[string]State)
	go loop()
	return nil
}

func runCmd(prefix string, f config.File) State {
	cmd := exec.Command(f.Cmd, f.Args)

	re, e := cmd.StderrPipe()
	if e != nil {
		return State{Err: e}
	}

	ro, e := cmd.StdoutPipe()
	if e != nil {
		return State{Err: e}
	}

	if config.Verbose {
		fmt.Printf("%s Run %s %s\n", prefix, f.Cmd, f.Args)
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

	cmdErr := cmd.Wait()
	ok := bytes.HasPrefix(stdOut, []byte("OK"))
	return State{
		Ok: ok,
		Stdout: string(stdOut),
		Stderr: string(stdErr),
		Err: cmdErr,
	}
}

func Check() {
	for name, f := range config.C.Files {
		if config.Verbose {
			fmt.Println("Check " + name)
		}

		prefix := fmt.Sprintf("worker(%s)", name)
		state := runCmd(prefix, f)
		state.prefix = name
		States[name] = state
		if config.Verbose {
			fmt.Printf("worker(%s) %+v\n", name, States[name])
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