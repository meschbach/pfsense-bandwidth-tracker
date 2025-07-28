package engine

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/melbahja/goph"
	"github.com/meschbach/pfsense-bandwidth-tracker/pkg/iftop"
	"golang.org/x/crypto/ssh"
	"io"
)

type Config struct {
	PfsenseUser      string
	PfsenseAddress   string
	PfsensePassword  string
	NetworkInterface string
}

func Run(ctx context.Context, config *Config, onFrameDone iftop.OnFrameDone) (problem error) {
	client, err := goph.NewConn(&goph.Config{
		User:     config.PfsenseUser,
		Addr:     config.PfsenseAddress,
		Port:     22,
		Auth:     goph.Password(config.PfsensePassword),
		Timeout:  goph.DefaultTimeout,
		Callback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		return err
	}
	defer func() {
		if err := client.Close(); err != nil {
			problem = errors.Join(err, problem)
		}
	}()
	eofTolerate := func(handle io.Closer) func() {
		return func() {
			if err := handle.Close(); err != nil {
				if !errors.Is(err, io.EOF) {
					problem = errors.Join(err, problem)
				}
			}
		}
	}
	fmt.Printf("Connected.\n")
	cmd, err := client.Command(fmt.Sprintf("iftop -nNbB -i %s -t -L 100 -P", config.NetworkInterface))
	if err != nil {
		return err
	}
	defer eofTolerate(cmd)

	fmt.Print("Waiting for command...\n")
	stdin, stdInErr := cmd.StdinPipe()
	stdout, stdOutErr := cmd.StdoutPipe()
	if err := errors.Join(stdInErr, stdOutErr); err != nil {
		return err
	}
	defer eofTolerate(stdin)

	if err := cmd.Start(); err != nil {
		return err
	}

	i := iftop.NewInterpreter(onFrameDone)
	lineSync := make(chan string, 256)
	go func() {
		defer close(lineSync)
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			lineSync <- line
		}
	}()

	for line := range lineSync {
		if err := i.Interpret(line); err != nil {
			return err
		}
	}
	return nil
}
