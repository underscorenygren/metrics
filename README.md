# metrics
Experimentation with metrics related stuff in golang


### Concepts

#### Source
Events originates at sources. They can be streaming,
such as reading from a file [file.go](pkg/file/file.go), or
event-based, such as events received as a webserver [http.go](pkg/http/http.go).

#### Sink

Sink is a destination for events, that emit no events itself. Events can
be discarded usining [blackhole.go](pkg/blackhole/blackhole.go), buffered
in memory for testing purposes [buffer.go](pkg/buffer/buffer.go) or written
to a persistent store like AWS kinesis [kinesis.go](pkg/kinesis/kinesis.go).

#### Stage

Stages connect sources and sinks, to allow events to flow. The simplest
stage [pipe.go](pkg/pipe/pipe.go) simply sends events from a source to a sink,
while [junction.go](pkg/junction/junction.go) routes events to different sinks
and [transformer.go](pkg/transformer/transformer.go) transforms events as they pass
through.

### Naming conventions
All Sources, Stages and Sinks are in the [./pkg](./pkg) directory,
in a package of their own, where the entrypoint is `nameofpkg.go`. Tests
live alongside the package.

Examples for packages are in [./examples](./examples), in a main packaged
keyed under their name.

For example, the `http` source, which receives events as a webserver, is
declared in [pkg/http/http.go](), with tests at [pkg/http/http_test.go]() and
an example at [examples/http/main.go]().

### Logging

We use a global logging module approach, to make it easy
to configure in tests, without putting any burden on
people instantiating modules having to know about
the internals of logging.

This does depend on a global module, which I'm a bit iffy
about. If I see a better approach to handle this, I'd
consider using that instead.


### Testing

Uses Ginkgo and Gomega for testing, [documented here](https://onsi.github.io/ginkgo/)


