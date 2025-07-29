package netstat

import (
	"context"
	"fmt"
	"github.com/meschbach/pfsense-bandwidth-tracker/pkg/engine"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"os"
	"time"
)

type Config struct {
	PfsenseUser      string
	PfsenseAddress   string
	PfsensePassword  string
	NetworkInterface string
}

var lineReadingCounter = promauto.NewCounter(prometheus.CounterOpts{
	Subsystem: "netstat",
	Name:      "lines_read_count",
	Help:      "Instance wide count of number of lines read",
})
var remoteCommandsFinished = prometheus.NewCounter(prometheus.CounterOpts{
	Subsystem: "netstat",
	Name:      "remote_ssh_exec_completed_count",
	Help:      "Number of remote commands completed",
})

type Netstat struct {
	config *Config
}

func NewNetstat(cfg *Config) *Netstat {
	return &Netstat{
		config: cfg,
	}
}

type nicMetrics struct {
	ingressBytesTotal prometheus.Gauge
	egressBytesTotal  prometheus.Gauge
}

type addressMetrics struct {
	ingressBytesTotal prometheus.Gauge
	egressBytesTotal  prometheus.Gauge
}

type metricsService struct {
	networkInterfaces map[string]*nicMetrics
	addresses         map[string]*addressMetrics
}

func (m *metricsService) recordMetics(reading []*IFaceReading) error {
	for _, r := range reading {
		if _, ok := m.networkInterfaces[r.Name]; !ok {
			labels := prometheus.Labels{}
			labels["nic"] = r.Name
			m.networkInterfaces[r.Name] = &nicMetrics{
				ingressBytesTotal: promauto.NewGauge(prometheus.GaugeOpts{
					Subsystem:   "netstat",
					Name:        "nic_ingress_bytes_total",
					Help:        "Total bytes received on this interface",
					ConstLabels: labels,
				}),
				egressBytesTotal: promauto.NewGauge(prometheus.GaugeOpts{
					Subsystem:   "netstat",
					Name:        "nic_egress_bytes_total",
					Help:        "Total bytes sent on this interface",
					ConstLabels: labels,
				}),
			}
		}
		m.networkInterfaces[r.Name].ingressBytesTotal.Set(float64(r.IfaceStats.Ingress.Bytes))
		m.networkInterfaces[r.Name].egressBytesTotal.Set(float64(r.IfaceStats.Egress.Bytes))

		for _, a := range r.AddressReadings {
			if _, ok := m.addresses[a.Network]; !ok {
				labels := prometheus.Labels{}
				labels["network"] = a.Network
				labels["address"] = a.Address
				labels["nic"] = r.Name
				m.addresses[a.Network] = &addressMetrics{
					ingressBytesTotal: promauto.NewGauge(prometheus.GaugeOpts{
						Subsystem:   "netstat",
						Name:        "address_ingress_bytes_total",
						Help:        "Total bytes received on this address",
						ConstLabels: labels,
					}),
					egressBytesTotal: promauto.NewGauge(prometheus.GaugeOpts{
						Subsystem:   "netstat",
						Name:        "address_egress_bytes_total",
						Help:        "Total bytes sent on this address",
						ConstLabels: labels,
					}),
				}
			}
			m.addresses[a.Network].ingressBytesTotal.Set(float64(a.Reading.Ingress.Bytes))
			m.addresses[a.Network].egressBytesTotal.Set(float64(a.Reading.Egress.Bytes))
		}
	}
	return nil
}

func (n *Netstat) RunService(context context.Context) (problem error) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	s := &metricsService{
		networkInterfaces: map[string]*nicMetrics{},
		addresses:         map[string]*addressMetrics{},
	}
	for {
		select {
		case <-context.Done():
			return context.Err()
		case <-ticker.C:
			if err := n.Tick(s.recordMetics); err != nil {
				return err
			}
		}
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
			lineReadingCounter.Inc()
			if err := i.consumeLine(*line.Stdout); err != nil {
				return err
			}
		}
	}

	remoteCommandsFinished.Inc()
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
