# Changelog

## [[v0.4.0]](https://github.com/mlange-42/beecs-cli/compare/v0.3.0...v0.4.0)

### Breaking changes

- Upgrade to beecs v0.4.0, with a renamed parameter (#57)

## [[v0.3.0]](https://github.com/mlange-42/beecs-cli/compare/v0.2.0...v0.3.0)

### Features

- Provides parameter `params.Termination.OnExtinction` to terminate simulations on extinction of all bees (#55)
- Provides observer `Extinction` to report the tick of colony extinction (#55)

## [[v0.2.0]](https://github.com/mlange-42/beecs-cli/compare/v0.1.0...v0.2.0)

### Features

- Allows to set a super-seed for random seed generation in experiments (#16, #17)
- Seeds of individual runs in experiments are written to the parameter output (#16)
- Adds sub-command `init` to create input file templates (#22)
- Adds support for seasonal and "scripted" patch dynamics (#23)
- Adds support for weather/foraging period files (#24)
- Adds support for adding custom resources/parameters as well as systems (#25)
- Rework of the command line interface for a simpler syntax when using default file names (#29, #36)
- Allow for multiple runs without providing an experiment (#33)
- Random seed is stored in experiments, but can be overwritten via CLI (#34)
- Record start and end time of each run in parameters output file (#37)
- Add JSON property `Final` to table output to write rows on finalization only (#38)
- Provides more visualizations, like ECS monitor and resources and systems inspector (#39, #43)
- Adds a view for the visualization of patch resources, visits, and colony status (#42, #43)
- Adds line plots for a full plot per update, like foraging stats (#44)
- Adds CSV output for snapshot table observers (#45)
- Adds a command line option `--index` for selecting simulation runs of an experiment (#46)
- Allows to seed individual runs with `-x params.RandomSeed.Seed=123` (#47)

### Other

- Show plots only when a single run is performed (#19)
- Restructured code for use as a library/module (#26)
- Renames JSON property `Params` to `ObserverConfig` (#49)
- Move `CustomParams` to base module (#50)

## [[v0.1.0]](https://github.com/mlange-42/beecs-cli/tree/v0.1.0)

Initial release of beecs-cli.
