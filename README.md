# beecs-cli

Command line interface for the [beecs](https://github.com/mlange-42/beecs) honeybee model.

## Installation

There are currently no precompiled binaries provided.

Install beecs-cli using [Go](https://go.dev)

```
go install github.com/mlange-42/beecs-cli@latest
```

## Usage

Get CLI help like this:

```
beecs-cli
```

A single, slowed down run od the base example:

```
beecs-cli -s 30 -p _examples/base/parameters.json -o _examples/base/observers.json
```

Run the full base example with 10 runs per parameter set:

```
beecs-cli -r 10 -p _examples/base/parameters.json -o _examples/base/observers.json -e _examples/base/experiment.json
```

See the [examples](https://github.com/mlange-42/beecs-cli/tree/main/_examples) for the format of the required JSON files.