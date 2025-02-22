package main

import "github.com/grafana/pyroscope-go"

func startProfiler() (*pyroscope.Profiler, error) {
	return pyroscope.Start(pyroscope.Config{
		ApplicationName: "todo-service",
		ServerAddress:   "http://pyroscope.local",
		Logger:          pyroscope.StandardLogger,
	})
}
