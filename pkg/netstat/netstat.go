package netstat

import (
	"context"
	"fmt"
	"github.com/meschbach/pfsense-bandwidth-tracker/pkg/engine"
	"os"
)

type Config struct {
	PfsenseUser      string
	PfsenseAddress   string
	PfsensePassword  string
	NetworkInterface string
}

type Netstat struct {
	config *Config
}

func NewNetstat(cfg *Config) *Netstat {
	return &Netstat{
		config: cfg,
	}
}

type OnReading func(reading []*IFaceReading) error

func (n *Netstat) Tick(onReading OnReading) (problem error) {
	c := engine.SSHStream{Config: &engine.Config{
		PfsenseUser:      n.config.PfsenseUser,
		PfsenseAddress:   n.config.PfsenseAddress,
		PfsensePassword:  n.config.PfsensePassword,
		NetworkInterface: n.config.NetworkInterface,
	}}
	lines, err := c.StreamCommand(256, "netstat", "-ibdnW")
	if err != nil {
		return err
	}

	i := &interpreter{
		state:        0,
		readings:     nil,
		currentIFace: nil,
	}
	for line := range lines {
		if line.Problem != nil {
			return line.Problem
		}
		if line.Stderr != nil {
			fmt.Fprintf(os.Stderr, "netstat.stderr(remote): %s\n", *line.Stderr)
		}
		if line.Stdout != nil {
			if err := i.consumeLine(*line.Stdout); err != nil {
				return err
			}
		}
	}

	result, err := i.done()
	if err != nil {
		return err
	}
	return onReading(result)
}

func (n *Netstat) TextUIOnce(ctx context.Context) (problem error) {
	return n.Tick(func(result []*IFaceReading) error {
		var addresses []struct {
			iface string
			addr  *AddressReading
		}
		fmt.Printf("%10s\t%7s\t%9s\t%9s\t%9s\n", "iface", "mtu", "collisons", "ingress", "egress")
		for _, r := range result {
			fmt.Printf("%10s\t%7d\t%9d\t%s\t%s\n", r.Name, r.MTU, r.Collisions, bytesToScaled(r.IfaceStats.Ingress.Bytes), bytesToScaled(r.IfaceStats.Egress.Bytes))
			for _, a := range r.AddressReadings {
				addresses = append(addresses, struct {
					iface string
					addr  *AddressReading
				}{
					iface: r.Name,
					addr:  a,
				})
			}
		}
		fmt.Printf("%20s\t%32s\t%10s\t%9s\t%9s\t%9s\t%9s\n", "network", "addr", "iface", "iPkts", "oPkts", "ingress", "egress")
		for _, s := range addresses {
			r := s.addr
			fmt.Printf("%20s\t%32s\t%10s\t%9s\t%9s\t%s\t%s\n", r.Network, r.Address, s.iface, numberToScaled(r.Reading.Ingress.Packets), numberToScaled(r.Reading.Ingress.Bytes), bytesToScaled(r.Reading.Ingress.Bytes), bytesToScaled(r.Reading.Egress.Bytes))
		}
		return nil
	})
}
