# go-otel-sample-app

OTel sample app in Go for demonstrating logs/traces/metrics

## TODO
- move otel.local and pyroscope.local to env vars
- create method to instantiate logger, tracer, meter providers
- move service names, package, any other magic string to main.go
- add console exporters under env var?
- screenshots (log, trace, metric, service map)
- add error log in client in case status code is not okay