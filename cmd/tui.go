package main

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

func tuiMain(config *options) error {
	client, err := goph.NewConn(&goph.Config{
		User:     config.pfsenseUser,
		Addr:     config.pfsenseAddress,
		Port:     22,
		Auth:     goph.Password(config.pfsensePassword),
		Timeout:  goph.DefaultTimeout,
		Callback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		return err
	}
	defer func() {
		if err := client.Close(); err != nil {
			panic(err)
		}
	}()
	eofTolerate := func(handle io.Closer) func() {
		return func() {
			if err := handle.Close(); err != nil {
				if !errors.Is(err, io.EOF) {
					panic(err)
				}
			}
		}
	}
	fmt.Printf("Connected.\n")
	cmd, err := client.Command(fmt.Sprintf("iftop -nNbB -i %s -t -L 100 -P", config.networkInterface))
	if err != nil {
		panic(err)
	}
	defer eofTolerate(cmd)

	fmt.Print("Waiting for command...\n")
	stdin, stdInErr := cmd.StdinPipe()
	stdout, stdOutErr := cmd.StdoutPipe()
	if err := errors.Join(stdInErr, stdOutErr); err != nil {
		panic(err)
	}
	defer eofTolerate(stdin)

	if err := cmd.Start(); err != nil {
		return err
	}

	i := iftop.NewInterpreter(func(ctx context.Context, reading *iftop.Reading, interpreter *iftop.IftopInterpreter) error {
		for _, f := range reading.Frames {
			fmt.Printf("\t%s\t%s\t<=>\t%s\t%s\n", f.Source.Address, f.Source.Cumulative, f.Destination.Address, f.Destination.Cumulative)
		}
		fmt.Printf("\n")
		return nil
	})
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
