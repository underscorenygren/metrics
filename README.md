# partaj

Package partaj is a minimal event stream processing framework for reading, processing and writing events,
using a DSL inspired by functional programming and more cumbersome event processing systems such as Apache Kafka.

## Why `partaj`?

There are many event processing systems out there, from cumbersome heavy-hitters like Kafka to intricate
functional languages and DSL custom-designed for your use case.

Partaj is attempting to bring some of that goodness to `golang`.
It proces a minimal set of concepts and tools taken from robust event streaming systems, to allow you
to define complex pipelines using a simple API, while leveraging some of the power of go's excellent
capabilities for paralellization and performance.

### Yes, but what does it mean?

_Partaj_ is Swedish slang for a party, and since parties make for great events, `partaj` was born.

## Installation

`go get -u github.com/underscorenygren/partaj`

`partaj` is written as a go 1.12 module.

### Installing binaries from source

If `go get` doesn't work for you to get the bineries, you can build them from source with:

`make install`

This will but binaries in `$GOPATH/bin`.

## Binaries

- `partaj-tail`
  - A cloudwatch log tailer. Use `-h` for usage

## Example Usage

Stands up a server and sends body of requests to it to AWS Firehose.
```go
package main

import (
  "github.com/underscorenygren/partaj/pkg/firehose"
  "github.com/underscorenygren/partaj/pkg/http"
  "os"
  "log"
)

func main() {
  sink, _ := firehose.NewSink("name-of-my-firehose")
  s, _ := http.NewServer(http.Config{
    Sink: sink,
  })
  log.Fatal(s.ListenAndServe())
}
```

## Core Concepts

At it's core, `partaj` consistes of three basic types of operators:
- `Sources`: where events originate
- `Sinks`: where events end up
- `Stages`: that connects sources and sinks.

An event processing system is called a Pipeline, and consists of one
or more stages connected together, with events flowing through them.

### Source

Events originates at sources. They can be streaming,
such as reading from a file [file.go](./pkg/file/file.go), or
event-based, such as events received as a webserver [http.go](./pkg/http/http.go).

### Sink

Sink is an event destination that emit no events itself. Events can
be discarded with [blackhole.go](./pkg/blackhole/blackhole.go), buffered
in memory for testing purposes [buffer.go](./pkg/buffer/buffer.go) or written
to a persistent store like AWS kinesis [kinesis.go](./pkg/kinesis/kinesis.go).

### Stage

Stages connect sources and sinks, to allow events to flow. The simplest
stage [pipe.go](pkg/pipe/pipe.go) simply sends events from a source to a sink.

## Documentation

Documentation can be viewed by running a godoc instance using `make docs`.

As of this writing, `go modules` and `godoc` don't yet play nicely, so we use docker to
make this more straightforward.

type `make open-docs` to navigate to the package page (unix-like only).

## Development

type `make` to see what build commands are available.

`make install-deps` and `make install-test` will install project prerequisites, such as localstack.

After installing deps, run

`make test`

and make sure they pass.

By default, localstack dependent tests will be skipped if localstack isn't running.

To test aws integrations, run `make localstack`, to start localstack if you don't run it already.

### Conventions

All Sources, Stages and Sinks are in the [./pkg](./pkg) directory,
in packages of their own, where the entrypoint is `nameofpkg.go`. Tests
live alongside their package.

Examples for packages are in [./examples](./examples), in a main package
keyed under their name.

For example, the `http` source, which receives events as a webserver, is
declared in [pkg/http/http.go](./pkg/http/http.go), with tests
at [pkg/http/http_test.go](./pkg/http/http_test.go) and
an example at [examples/http/main.go](./examples/http/main.go).

Executable binaries are in the [./cmd](./cmd) directory.

### Logging

`partaj` uses a global logging module approach, so that logging can be
configured globally in tests, and there is no burden on users of `partaj`
to know about the internal logging.

Internally, it uses the `zap` logging framework, which takes care of optimizing
away slow calls when writing Debug events.

### Testing

Uses Ginkgo and Gomega for testing, [documented here](https://onsi.github.io/ginkgo/).
