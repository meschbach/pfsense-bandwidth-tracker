package main

import (
	"context"
	"fmt"
	"github.com/meschbach/pfsense-bandwidth-tracker/pkg/engine"
	"github.com/meschbach/pfsense-bandwidth-tracker/pkg/iftop"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

var frames = promauto.NewCounter(prometheus.CounterOpts{
	Name: "iftop_readings",
	Help: "A full reading from the remote iftop instance",
})

func runService(config *options) error {
	go func() {
		gauges := make(map[string]prometheus.Gauge)
		err := engine.Run(context.Background(), &engine.Config{
			PfsenseUser:      config.pfsenseUser,
			PfsenseAddress:   config.pfsenseAddress,
			PfsensePassword:  config.pfsensePassword,
			NetworkInterface: config.networkInterface,
		}, func(ctx context.Context, reading *iftop.Reading, interpreter *iftop.IftopInterpreter) error {
			frames.Add(1)
			for _, f := range reading.Frames {
				k := f.Source.Address + f.Destination.Address
				if g, ok := gauges[k]; ok {
					g.Set(f.Source.Cumulative.ToFloat64())
				} else {
					labels := make(prometheus.Labels)
					source := f.Source.AddressParts()
					labels["src_host"] = source.Host
					labels["src_port"] = source.Port

					destination := f.Destination.AddressParts()
					labels["dst_host"] = destination.Host
					labels["dst_port"] = destination.Port
					labels["iface"] = config.networkInterface
					g := promauto.NewGauge(prometheus.GaugeOpts{
						Name:        "bandwidth",
						Help:        "in bytes",
						ConstLabels: labels,
					})
					gauges[k] = g
					g.Set(f.Source.Cumulative.ToFloat64())
				}
			}
			return nil
		})
		if err != nil {
			panic(err)
		}
	}()
	fmt.Printf("Exporting prometheus service on :2112\n")
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
	return nil
}
