package worker

import (
	"github.com/mpdroog/healthd/config"
	"time"
	"os/exec"
	"fmt"
	"io/ioutil"
	"bytes"
	"strings"
)

type State struct {
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
	return strings.Replace(msg, "\n", "", -1)
}

var States map[string]State

func Init() error {
	States = make(map[string]State)
	go loop()
	return nil
}

func runCmd(f config.File) State {
	cmd := exec.Command(f.Cmd, "")

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
	for _, f := range config.C.Files {
		cmd := f.Cmd
		if config.Verbose {
			fmt.Println("Check " + cmd)
		}

		prefix := fmt.Sprintf("worker(%s)", cmd)
		States[cmd] = runCmd(f)
		if config.Verbose {
			fmt.Printf("%s %+v\n", prefix, States[cmd])
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