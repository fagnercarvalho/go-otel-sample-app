package main

import "github.com/grafana/pyroscope-go"

func startProfiler(address, serviceName string) (*pyroscope.Profiler, error) {
	return pyroscope.Start(pyroscope.Config{
		ApplicationName: serviceName,
		ServerAddress:   address,
		Logger:          pyroscope.StandardLogger,
	})
}
