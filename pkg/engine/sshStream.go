package engine

import (
	"bufio"
	"errors"
	"github.com/melbahja/goph"
	"golang.org/x/crypto/ssh"
	"io"
)

type SSHStream struct {
	Config *Config
}

type SSHStreamLine struct {
	Stdout  *string
	Stderr  *string
	Problem error
}

func (s *SSHStream) StreamCommand(bufferSize int, program string, args ...string) (output <-chan SSHStreamLine, problem error) {
	sync := make(chan SSHStreamLine, bufferSize)
	go func() {
		defer close(sync)
		run := func() (problem error) {
			eofTolerate := func(handle io.Closer) func() {
				return func() {
					if err := handle.Close(); err != nil {
						if !errors.Is(err, io.EOF) {
							problem = errors.Join(err, problem)
						}
					}
				}
			}

			client, err := goph.NewConn(&goph.Config{
				User:     s.Config.PfsenseUser,
				Addr:     s.Config.PfsenseAddress,
				Port:     22,
				Auth:     goph.Password(s.Config.PfsensePassword),
				Timeout:  goph.DefaultTimeout,
				Callback: ssh.InsecureIgnoreHostKey(),
			})
			if err != nil {
				return err
			}
			defer eofTolerate(client)

			cmd, err := client.Command(program, args...)
			if err != nil {
				return err
			}
			defer eofTolerate(cmd)

			stdin, stdInErr := cmd.StdinPipe()
			stdout, stdOutErr := cmd.StdoutPipe()
			stderr, stdErrErr := cmd.StderrPipe()
			if err := errors.Join(stdInErr, stdOutErr, stdErrErr); err != nil {
				return err
			}
			defer eofTolerate(stdin)

			if err := cmd.Start(); err != nil {
				return err
			}

			go func() {
				scanner := bufio.NewScanner(stderr)
				for scanner.Scan() {
					line := scanner.Text()
					sync <- SSHStreamLine{Stderr: &line}
				}
			}()

			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				line := scanner.Text()
				sync <- SSHStreamLine{
					Stdout: &line,
				}
			}
			return scanner.Err()
		}
		problem := run()
		if problem != nil {
			sync <- SSHStreamLine{
				Problem: problem,
			}
		}
	}()
	return sync, nil
}
