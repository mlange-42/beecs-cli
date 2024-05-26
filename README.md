# beecs-cli

[![Test status](https://img.shields.io/github/actions/workflow/status/mlange-42/beecs-cli/tests.yml?branch=main&label=Tests&logo=github)](https://github.com/mlange-42/beecs-cli/actions/workflows/tests.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/mlange-42/beecs-cli)](https://goreportcard.com/report/github.com/mlange-42/beecs-cli)
[![Go Reference](https://img.shields.io/badge/reference-%23007D9C?logo=go&logoColor=white&labelColor=gray)](https://pkg.go.dev/github.com/mlange-42/beecs-cli)
[![GitHub](https://img.shields.io/badge/github-repo-blue?logo=github)](https://github.com/mlange-42/beecs-cli)

Command line interface for the [beecs](https://github.com/mlange-42/beecs) honeybee model and derivatives.

## Features

* Model parametrization from JSON files.
* Parallelized simulation experiments, with parameter variation configured via JSON.
* 100% reproducible simulation experiments.
* Various real-time visualizations for debugging and model exploration.
* All this, even for derived models with altered sub-models, state variables etc.

For a graphical user interface for the beecs model, see [beecs-ui](https://github.com/mlange-42/beecs-ui).

## Installation

Pre-compiled binaries for Linux, Windows and MacOS are available in the [Releases](https://github.com/mlange-42/beecs-cli/releases).

> To install the latest **development version** using [Go](https://go.dev), run:
> 
> ```
> go install github.com/mlange-42/beecs-cli@main
> ```
> 
> Note: Use `beecs-cli` instead of `beecs` in the examples below in this case.

## CLI usage

Get CLI help like this:

```
beecs -h
```

A single simulation with live plots, at 30 ticks per second:

```
beecs -d _examples/base --observers --tps 30
```

Run the full base example with parameter variation and 10 runs per parameter set:

```
beecs -d _examples/base --observers --experiment -r 10
```

Print all default parameters in the tool's input format:

```
beecs parameters
```

Create input file templates in the current directory:

```
beecs init
```

## Library usage

With beecs-cli, it also is possible to fully parameterize models derived from the original [beecs](https://github.com/mlange-42/beecs) model,
although they use custom systems and/or parameters and global state variables.

An example for how to modify [beecs](https://github.com/mlange-42/beecs) while using beecs-cli is provided by the repository [beecs-template](https://github.com/mlange-42/beecs-template).


## Input files

All file locations are relative to the working directory given by `-d` (defaults to the current directory).

One or more **parameter files** are required. Entries in those files overwrite the default parameters.
The default is file `parameters.json` in the working directory. Here is an example:

```json
{
    "Parameters": {
        "Termination": {
            "MaxTicks": 365
        },
        "InitialPopulation": {
            "Count": 25000
        }
    }
}
```

For any kind of output, an **observers file** is required.
It specifies which observers for visualizations or file output should be used.
Here is an example:

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

Observers must be enabled using the `-o` flag. The default is file `observers.json` in the working directory. 

These files are sufficient for single simulations with visual of file output.

With a further **experiment file**, parameters can be systematically varied in various ways.
Here is an example:

```json
{
    "Seed": 123,
    "Parameters": [
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
}
```

Experiments must be enabled using the `-e` flag. The default is file `experiments.json` in the working directory.

> Note: The prefix `params.` is required to unambiguously identify the type of the parameter group to modify.

See also the [examples](https://github.com/mlange-42/beecs-cli/tree/main/_examples) for the format of the required JSON files.
