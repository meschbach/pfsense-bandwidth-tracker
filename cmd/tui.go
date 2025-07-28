package main

import (
	"context"
	"fmt"
	"github.com/meschbach/pfsense-bandwidth-tracker/pkg/engine"
	"github.com/meschbach/pfsense-bandwidth-tracker/pkg/iftop"
)

func tuiMain(config *options) error {
	return engine.Run(context.Background(), &engine.Config{
		PfsenseUser:      config.pfsenseUser,
		PfsenseAddress:   config.pfsenseAddress,
		PfsensePassword:  config.pfsensePassword,
		NetworkInterface: config.networkInterface,
	}, func(ctx context.Context, reading *iftop.Reading, interpreter *iftop.IftopInterpreter) error {
		for _, f := range reading.Frames {
			fmt.Printf("\t%s\t%s\t<=>\t%s\t%s\n", f.Source.Address, f.Source.Cumulative, f.Destination.Address, f.Destination.Cumulative)
		}
		fmt.Printf("\n")
		return nil
	})
}
