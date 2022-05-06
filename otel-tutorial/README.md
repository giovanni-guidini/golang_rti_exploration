This folder contains basically the example app from OpenTelemetry [Getting Started guide in Go](https://opentelemetry.io/docs/instrumentation/go/getting-started/).
That was modified to try and add coverage info to it.

`opentelem/opentelem.go` contains the open telemetry related implementations.
The interesting thing is that it won't work, because opentelem is in a different package from the rest of the app.
So it doesn't has access to the package-wide GoCover struct, and there's no call to the RecordCoverageMap function from the main package.

As a workaround we manually send GoCover info to the SPanProcessor via `opentelem.RecordCoverageMap`. To do that uncomment lines 15 and 56 in `annotated_app.go`.
Needless to say that this is a strong hack.

`annotated_app.go` is the result of running the modified cover tool (`../cover/cover_copy.go`) in the `app.bak` file.