package worker

import (
	"context"
	"testing"
	"time"
)

func TestRunTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	s := runCmd(ctx, "../script.unittest/timeout.sh")
	if s.String() != "signal: killed" {
		t.Errorf("timeout-script other output than expected output=%s", s.String())
	}
}

func TestRunInvalid(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	s := runCmd(ctx, "../script.unittest/invalidtest.sh")
	if s.String() != "Stdout not OK" {
		t.Errorf("invalidtest-script should error, output=%s", s.String())
	}
}
