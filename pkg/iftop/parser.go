package iftop

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
)

type iftopState int

const (
	iftopStart iftopState = iota
	iftopframeHeaderStart
	iftopframeHeaderEnd
	iftopFirstLine
	iftopSecondLine
	iftopMaybeLine
	iftopFooter
	iftopFooterEnd
)

type AddressParts struct {
	Host string
	Port string
}

type BandwidthDirection struct {
	Address    string
	Cumulative ByteReading
}

func (b BandwidthDirection) AddressParts() AddressParts {
	host, port, err := net.SplitHostPort(b.Address)
	if err != nil {
		return AddressParts{Host: b.Address, Port: ""}
	}
	return AddressParts{Host: host, Port: port}
}

type Frame struct {
	Index       int
	Source      BandwidthDirection
	Destination BandwidthDirection
}

type Reading struct {
	Frames []*Frame
}

type OnFrameDone func(ctx context.Context, reading *Reading, interpreter *IftopInterpreter) error

type IftopInterpreter struct {
	state          iftopState
	nicName        string
	currentFrame   *Frame
	currentReading *Reading
	onFrameDone    OnFrameDone
}

func NewInterpreter(onFrameDone OnFrameDone) *IftopInterpreter {
	return &IftopInterpreter{
		state:       iftopStart,
		onFrameDone: onFrameDone,
	}
}

func (i *IftopInterpreter) Interpret(line string) error {
	if strings.Trim(line, " ") == "" {
		return nil
	}
	switch i.state {
	case iftopStart:
		return i.consumeStart(line)
	case iftopframeHeaderStart:
		i.state = iftopframeHeaderEnd
		return nil
	case iftopframeHeaderEnd:
		if line[0] == '-' {
			i.currentReading = &Reading{}
			i.state = iftopFirstLine
		}
		return nil
	case iftopFirstLine:
		return i.consumeFirstLine(line)
	case iftopSecondLine:
		return i.consumeSecondLine(line)
	case iftopMaybeLine:
		if line[0] == '-' {
			i.state = iftopFooter
			return nil
		} else {
			return i.consumeFirstLine(line)
		}
	case iftopFooter:
		if line[0] == '=' {
			i.state = iftopframeHeaderEnd
			return i.onFrameDone(context.Background(), i.currentReading, i)
		}
		return nil
	case iftopFooterEnd:
		i.state = iftopframeHeaderStart
		return nil
	default:
		panic(fmt.Sprintf("Unexpected state %d\n", i.state))
	}
}

func (i *IftopInterpreter) consumeStart(line string) error {
	args, err := fmt.Sscanf(line, "Listening on %s\n", &i.nicName)
	if err != nil {
		return err
	}
	if args != 1 {
		return errors.New(fmt.Sprintf("Expected 1 argument, got %d\n", args))
	}
	fmt.Printf("** Intefface %s\n", i.nicName)
	i.state = iftopframeHeaderStart
	return nil
}

func (i *IftopInterpreter) consumeFirstLine(line string) error {
	i.currentFrame = &Frame{}
	var last2, last10, last40 string
	args, err := fmt.Sscanf(line, "%d\t%40s\t=>\t%10s\t%10s\t%10s\t%10s", &i.currentFrame.Index, &i.currentFrame.Source.Address, &last2, &last10, &last40, &i.currentFrame.Source.Cumulative)
	if err != nil {
		return err
	}
	if args != 6 {
		return errors.New(fmt.Sprintf("Expected 1 argument, got %d\n", args))
	}
	i.state = iftopSecondLine
	return nil
}

func (i *IftopInterpreter) consumeSecondLine(line string) error {
	var last2, last10, last40 string
	args, err := fmt.Sscanf(line, "%s\t<=\t%10s\t%10s\t%10s\t%10s", &i.currentFrame.Destination.Address, &last2, &last10, &last40, &i.currentFrame.Destination.Cumulative)
	if err != nil {
		return err
	}
	if args != 5 {
		return errors.New(fmt.Sprintf("Expected 1 argument, got %d\n", args))
	}
	i.currentReading.Frames = append(i.currentReading.Frames, i.currentFrame)
	i.currentFrame = nil
	i.state = iftopMaybeLine
	return nil
}
