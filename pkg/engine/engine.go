package engine

import (
	"context"
	"fmt"
	"github.com/meschbach/pfsense-bandwidth-tracker/pkg/iftop"
	"os"
)

type Config struct {
	PfsenseUser      string
	PfsenseAddress   string
	PfsensePassword  string
	NetworkInterface string
}

func Run(ctx context.Context, config *Config, onFrameDone iftop.OnFrameDone) (problem error) {
	streamer := SSHStream{Config: config}
	lines, err := streamer.StreamCommand(256, "iftop", "-nNbB", "-i", config.NetworkInterface, "-t", "-L", "100", "-P")
	if err != nil {
		return err
	}

	i := iftop.NewInterpreter(onFrameDone)
	for l := range lines {
		if l.Problem != nil {
			return l.Problem
		}
		if l.Stdout != nil {
			if err := i.Interpret(*l.Stdout); err != nil {
				return err
			}
		}
		if l.Stderr != nil {
			fmt.Fprintf(os.Stderr, "iftop.stderr(remote): %s\n", *l.Stderr)
		}
	}
	return nil
}
