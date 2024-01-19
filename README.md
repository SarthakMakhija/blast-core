<p align="center">
	<img alt="blast" src="https://github.com/SarthakMakhija/blast-core/assets/21108320/5c55527f-ece4-478f-b6a3-f26c536c232a" />
</p>


| Platform      | Build Status                                                                                                                                                                                                       |
|---------------|:-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Ubuntu latest | [![blast_ubuntu_latest](https://github.com/SarthakMakhija/blast-core/actions/workflows/build_ubuntu_latest.yml/badge.svg)](https://github.com/SarthakMakhija/blast-core/actions/workflows/build_ubuntu_latest.yml) |
| macOS 12      | [![blast_macos_12](https://github.com/SarthakMakhija/blast-core/actions/workflows/build_macos_12.yml/badge.svg)](https://github.com/SarthakMakhija/blast-core/actions/workflows/build_macos_12.yml)                |


**blast** is a load generator for TCP servers, especially if such servers maintain persistent connections.

## Content Organization

- [Why blast](#why-blast)
- [blast-core](#blast-core)
- [Features](#features)
- [FAQs](#faqs)
- [References](#references)

## Why blast

I am a part of the team that is developing a strongly consistent distributed key/value storage engine with support for rich queries.
The distributed key/value storage engine has TCP servers that implement [Single Socket Channel](https://martinfowler.com/articles/patterns-of-distributed-systems/single-socket-channel.html) and [Request Pipeline
](https://martinfowler.com/articles/patterns-of-distributed-systems/request-pipeline.html). 

We needed a way to send load on our servers and get a report with details including total connections established, total requests sent, total responses read and time to get those responses back etc.

Another detail, our servers accept protobuf encoded messages as byte slices, so the tool should be able to send the load (/byte slice) in a format that the target servers
can decode. Almost all distributed systems accept payloads in a very specific format. For example, [JunoDB](https://github.com/paypal/junodb) sends (and receives) [OperationalMessage](https://github.com/paypal/junodb/blob/ca68aa14734768fd047b66ea0b7e6316b15fef16/pkg/proto/opMsg.go#L33) encoded as byte slice.

All we needed was a tool that can send load (or specific load) on target TCP servers, read responses from those servers and present a decent :) report. This was an opportunity to build **blast**. **blast** is inspired from [hey](https://github.com/rakyll/hey), which is an HTTP load generator in golang.

## blast-core

**blast-core** is the heart of [blast CLI](https://github.com/SarthakMakhija/blast). It provides support for sending requests, reading payload from file, reading responses and generating reports.
It also provides support parsing command line arguments to simplify building a custom CLI. More on this in the [FAQs](#faqs).  

## Features

**blast-core** provides the following features:
1. Support for **sending N requests** to the target server.
2. Support for **reading N total responses** from the target server.
3. Support for **reading N successful responses** from the target server.
4. Support for **customizing** the **load** **duration**. By default, blast runs for 20 seconds.
5. Support for sending N requests to the target server with the specified **concurrency** **level**.
6. Support for **establishing N connections** to the target server.
7. Support for specifying the **connection timeout**.
8. Support for specifying **requests per second** (also called **throttle**).
9. Support for **printing** the **report**.
10. Support for sending dynamic payloads with **PayloadGenerator**.

## FAQs

1. **How does blast-core support dynamic payload?**

**blast-core** provides an interface called `PayloadGenerator` and ships with an implementation called `ConstantPayloadGenerator`. Anyone can implement
the interface `PayloadGenerator` to generate dynamic payload.

2. **Does blast (CLI) provide support for dynamic payload?**

No, [blast CLI](https://github.com/SarthakMakhija/blast) does not provide support for dynamic payload. To implement dynamic payload, one needs to create their
own CLI. **blast-core** makes it easy to build a thin CLI. Creating a custom CLI involves the following:

  - Express the dependency on **blast-core** in `go.mod`
  - Write a thin `main` function that provides the implementation of `PayloadGenerator`

The `main` looks like the following:

```go
func main() {
    commandArguments := blast.NewCommandArguments()
    blastInstance := commandArguments.ParseWithDynamicPayload(executableName, <<An implementation of PayloadGenerator>>)

    interruptChannel := make(chan os.Signal, 1)
    signal.Notify(interruptChannel, os.Interrupt)

    go func() {
        <-interruptChannel
        blastInstance.Stop()
    }()

    blastInstance.WaitForCompletion()
}
```

## References
[hey](https://github.com/rakyll/hey)

*The logo is built using [logo.com](logo.com)*.
