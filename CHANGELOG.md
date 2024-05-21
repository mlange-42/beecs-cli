## [[unpublished]](https://github.com/mlange-42/beecs-cli/compare/v0.1.0...main)

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

### Other

- Show plots only when a single run is performed (#19)
- Restructured code for use as a library/module (#26)

## [[v0.1.0]](https://github.com/mlange-42/beecs-cli/tree/v0.1.0)

Initial release of beecs-cli.
