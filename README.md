# beecs-cli

[![Test status](https://img.shields.io/github/actions/workflow/status/mlange-42/beecs-cli/tests.yml?branch=main&label=Tests&logo=github)](https://github.com/mlange-42/beecs-cli/actions/workflows/tests.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/mlange-42/beecs-cli)](https://goreportcard.com/report/github.com/mlange-42/beecs-cli)
[![GitHub](https://img.shields.io/badge/github-repo-blue?logo=github)](https://github.com/mlange-42/beecs-cli)

Command line interface for the [beecs](https://github.com/mlange-42/beecs) honeybee model.

## Purpose

* Model parametrization from JSON files.
* Systematic, parallel simulations with parameter variation, configured with JSON.
* Running single simulations with various real-time visualizations.

## Installation

Pre-compiled binaries for Linux, Windows and MacOS are available in the [Releases](https://github.com/mlange-42/beecs-cli/releases).

To install the latest development version using [Go](https://go.dev), run:

```
go install github.com/mlange-42/beecs-cli@main
```

## Usage

Get CLI help like this:

```
beecs-cli -h
```

A single, slowed down run of the base example, with live plots:

```
beecs-cli -s 30 -d _examples/base
```

Run the full base example with parameter variation and 10 runs per parameter set:

```
beecs-cli -r 10 -d _examples/base -e experiment.json
```

Print all default parameters in the tool's input format:

```
beecs-cli parameters
```

Create input file templates in the current directory:

```
beecs-cli init
```

### Input files

All file locations are relative to the working directory given by `-d` (defaults to the current directory).

One or more **parameter files** are required. Entries in those files overwrite the default parameters.
The default is file `parameters.json` in the working directory. Here is an example:

```json
{
    "Termination": {
        "MaxTicks": 365
    },
    "InitialPopulation": {
        "Count": 25000
    }
}
```

For any kind of output, an **observers file** is required.
It specifies which observers for visualizations or file output should be used.
The default is file `parameters.json` in the working directory. Here is an example:

```json
{
    "Parameters": "out/Parameters.csv",
    "Tables": [
        {
            "Observer": "obs.WorkerCohorts",
            "File": "out/WorkerCohorts.csv"
        }
    ],
}
```

These files are sufficient for single simulations with visual of file output.

With a further **experiment file**, parameters can be systematically varied in various ways.
Here is an example:

```json
[
    {
        "Parameter": "params.Nursing.MaxBroodNurseRatio",
        "SequenceFloatRange": {
            "Min": 2.0,
            "Max": 4.0,
            "Values": 11
        }
    },
    {
        "Parameter": "params.Nursing.ForagerNursingContribution",
        "SequenceFloatValues": {
            "Values": [0, 0.25, 0.5]
        }
    }
]
```

> Note: The prefix `params.` is required to unambiguously identify the type of the parameter group to modify.

See also the [examples](https://github.com/mlange-42/beecs-cli/tree/main/_examples) for the format of the required JSON files.
