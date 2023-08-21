<p align="center">
    <img alt="blast" src="https://github.com/SarthakMakhija/blast/assets/21108320/ffb3336c-688f-4b33-b99f-3a26fb35982d" />
</p>

| Platform       | Build Status                                                                                                                  |
|----------------|:------------------------------------------------------------------------------------------------------------------------------|
| Ubuntu latest  | [![blast_ubuntu_latest](https://github.com/SarthakMakhija/blast/actions/workflows/build_ubuntu_latest.yml/badge.svg)](https://github.com/SarthakMakhija/blast/actions/workflows/build_ubuntu_latest.yml)|
| macOS 12       | [![blast_macos_12](https://github.com/SarthakMakhija/blast/actions/workflows/build_macos_12.yml/badge.svg)](https://github.com/SarthakMakhija/blast/actions/workflows/build_macos_12.yml)|


**blast** is a load generator for TCP servers, especially if such servers maintain persistent connections.

## Content Organization

## Why blast

I am a part of the team that is developing a strongly consistent distributed key/value storage engine with support for rich queries.
The distributed key/value storage engine has TCP servers that implement [Single Socket Channel](https://martinfowler.com/articles/patterns-of-distributed-systems/single-socket-channel.html) and [Request Pipeline
](https://martinfowler.com/articles/patterns-of-distributed-systems/request-pipeline.html). 

We needed a way to send load on our servers and get a report with some details including total connections created, total requests sent, total responses read and time to get those responses back etc.

Our servers accept protobuf encoded messages as byte slices, so the tool should be able to send the load (/byte slice) in a format that the target servers
can decode. Almost all distributed systems accept payloads in a very specific format. For example, [JunoDB](https://github.com/paypal/junodb) sends (and receives) [OperationalMessage](https://github.com/paypal/junodb/blob/ca68aa14734768fd047b66ea0b7e6316b15fef16/pkg/proto/opMsg.go#L33) encoded as byte slice.

All we needed was a tool that can send load (or specific load) on target TCP servers, read responses from those servers and present a decent :) report. This was an opportunity to build **blast**.

## Features

**blast** provides the following features:
1. Support for **sending N requests** to the target server.
2. Support for **reading N total responses** from the target server.
3. Support for **reading N successful responses** from the target server.
4. Support for **customizing** the **load** **duration**. By default, blast runs for 20 seconds.
5. Support for sending N requests to the target server with specified **concurrency** **level**.
6. Support for **establishing N connections** to the target server.
7. Support for specifying the **connection timeout**.
8. Support for specifying **requests per second** (also called **throttle**).
9. Support for **printing** the **report**.

## Installation

## FAQs

## Screenshots

- **Sending load on the target server:** `./blast -n 200000 -c 100 -conn 100  -f ./payload localhost:8989`

  <img width="715" alt="Sending load on the target server" src="https://github.com/SarthakMakhija/blast/assets/21108320/5a614f53-31cc-43b3-99ad-0cdfd22603e6">
- **Reading responses from the target server:** `./blast -n 200000 -c 100 -conn 100  -f ./payload -Rr -Rrs 19 -Rtr 200000 localhost:8989`
  
  <img width="715" alt="Reading responses from the target server" src="https://github.com/SarthakMakhija/blast/assets/21108320/2b8f7abe-c9eb-4ce1-a95e-a3c926074062">

- **Error distribution:** `./blast -n 200000 -c 100 -conn 100  -f ./payload localhost:8989`

  <img width="715" alt="Error distribution" src="https://github.com/SarthakMakhija/blast/assets/21108320/808ab493-8f8b-4792-bce2-bc8ee49d1d63">

## References
[hey](https://github.com/rakyll/hey)

The logo is built using [logo.com](logo.com)
