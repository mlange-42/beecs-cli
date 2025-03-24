package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/mlange-42/ark-tools/app"
	"github.com/mlange-42/beecs-cli/internal/run"
	"github.com/mlange-42/beecs-cli/internal/util"
	"github.com/mlange-42/beecs/experiment"
	"github.com/mlange-42/beecs/params"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	parametersFile = "parameters.json"
	observersFile  = "observers.json"
	experimentFile = "experiment.json"
	systemsFile    = "systems.json"
)

// Run the CLI app.
func Run() {
	if err := rootCommand().Execute(); err != nil {
		fmt.Printf("ERROR: %s\n", err.Error())
		fmt.Print("\nRun `beecs -h` for help!\n\n")
		os.Exit(1)
	}
}

// rootCommand sets up the CLI
func rootCommand() *cobra.Command {
	var dir string
	var outDir string
	var paramFiles []string
	var expFile string
	var obsFile string
	var sysFile string
	var speed float64
	var threads int
	var runs int
	var overwrite []string
	var seed int
	var indicesStr string

	var root cobra.Command
	root = cobra.Command{
		Use:           "beecs",
		Short:         "beecs-cli provides a command line interface for the beecs model.",
		Long:          `beecs-cli provides a command line interface for the beecs model.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(paramFiles) == 0 {
				_ = cmd.Help()
				os.Exit(0)
			}

			flagUsed := map[string]bool{}
			root.Flags().Visit(func(f *pflag.Flag) {
				flagUsed[f.Name] = true
			})

			rootRng := rand.New(rand.NewPCG(0, uint64(time.Now().UTC().Nanosecond())))

			if outDir == "" {
				outDir = dir
			}

			p := params.CustomParams{
				Parameters: params.Default(),
			}
			for _, f := range paramFiles {
				err := p.FromJSONFile(path.Join(dir, f))
				if err != nil {
					return err
				}
			}
			p.Parameters.WorkingDirectory.Path = dir

			var exp experiment.Experiment
			var rng *rand.Rand
			var err error
			if flagUsed["experiment"] {
				exp, rng, err = util.ExperimentFromFile(path.Join(dir, expFile), runs, seed)
				if err != nil {
					return err
				}
			} else {
				seedUsed := uint64(seed)
				if seed <= 0 {
					seedUsed = rootRng.Uint64()
				}
				rng = rand.New(rand.NewPCG(0, seedUsed))
				exp, err = experiment.New([]experiment.ParameterVariation{}, rng, runs)
				if err != nil {
					return err
				}
			}

			var observers util.ObserversDef
			if flagUsed["observers"] {
				observers, err = util.ObserversDefFromFile(path.Join(dir, obsFile))
				if err != nil {
					return err
				}
			}

			var systems []app.System
			if flagUsed["systems"] {
				systems, err = util.SystemsFromFile(path.Join(dir, sysFile))
				if err != nil {
					return err
				}
			}

			overwriteParams := make([]experiment.ParameterValue, len(overwrite))
			for i, s := range overwrite {
				parts := strings.Split(s, "=")
				if len(parts) != 2 {
					return fmt.Errorf("invalid syntax in option --overwrite (-x)")
				}
				overwriteParams[i] = experiment.ParameterValue{
					Parameter: parts[0],
					Value:     parts[1],
				}
			}

			indices, err := util.ParseIndices(indicesStr)
			if err != nil {
				return err
			}

			if exp.TotalRuns() <= 1 || len(indices) == 1 {
				threads = 1
			}
			if threads <= 1 {
				return run.Sequential(&p, &exp, &observers, systems, overwriteParams, outDir, speed, rng, indices)
			}
			return run.Parallel(&p, &exp, &observers, systems, overwriteParams, outDir, threads, speed, rng, indices)
		},
	}

	root.Flags().StringVarP(&dir, "directory", "d", ".", "Working directory")
	root.Flags().StringVarP(&outDir, "output", "", "", "Output directory if different from working directory")
	root.Flags().StringSliceVarP(&paramFiles, "parameters", "p", []string{parametersFile},
		"Parameter files, processed in the given order\n")

	root.Flags().StringVarP(&expFile, "experiment", "e", "",
		"Run experiment.\n Optionally, provide an experiment file for parameter variation")
	root.Flag("experiment").NoOptDefVal = experimentFile

	root.Flags().StringVarP(&obsFile, "observers", "o", "",
		"Run with observers.\n Optionally, provide an observers file for adding observers")
	root.Flag("observers").NoOptDefVal = observersFile

	root.Flags().StringVarP(&sysFile, "systems", "s", "",
		"Run with custom systems.\n Optionally, provide a systems file for using custom systems\n or changing the scheduling")
	root.Flag("systems").NoOptDefVal = systemsFile

	root.Flags().IntVarP(&seed, "seed", "", 0,
		"Overwrite experiment super random seed for seed generation.\n Default: don't overwrite.\n Use -1 to force random seeding")

	root.Flags().StringSliceVarP(&overwrite, "overwrite", "x", []string{}, "Overwrite variables like key1=value1,key2=value2")
	root.Flags().IntVarP(&runs, "runs", "r", 1, "Runs per parameter set")
	root.Flags().IntVarP(&threads, "threads", "t", runtime.NumCPU(), "Number of threads")
	root.Flags().Float64VarP(&speed, "tps", "", 0, "Speed limit in ticks per second. Default: 0 (unlimited)")
	root.Flags().StringVarP(&indicesStr, "index", "i", "", "Only run the given list or range of indices.\nExample: '2-5,8,12'. Default: all")

	root.Flags().SortFlags = false

	root.AddCommand(initCommand())
	root.AddCommand(parametersCommand())

	return &root
}

func parametersCommand() *cobra.Command {
	var dir string
	var paramFiles []string

	root := &cobra.Command{
		Use:           "parameters",
		Short:         "Prints all default model parameters in JSON format.",
		Long:          `Prints all default model parameters in JSON format.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			p := params.CustomParams{
				Parameters: params.Default(),
				Custom:     map[reflect.Type]any{},
			}
			for _, f := range paramFiles {
				err := p.FromJSONFile(path.Join(dir, f))
				if err != nil {
					return err
				}
			}

			js, err := p.ToJSON()
			if err != nil {
				return err
			}

			fmt.Println(string(js))

			return nil
		},
	}
	root.Flags().StringVarP(&dir, "directory", "d", ".", "Working directory")
	root.Flags().StringSliceVarP(&paramFiles, "parameters", "p", []string{}, "Optional parameter files, processed in the given order")

	return root
}

func initCommand() *cobra.Command {
	var dir string

	root := &cobra.Command{
		Use:           "init",
		Short:         "Initialize templates for an experiment.",
		Long:          `Initialize templates for an experiment.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			parFile := path.Join(dir, parametersFile)
			obsFile := path.Join(dir, observersFile)
			expFile := path.Join(dir, experimentFile)

			if fileExists(parFile) {
				return fmt.Errorf("parameter file '%s' already exists", parFile)
			}
			if fileExists(obsFile) {
				return fmt.Errorf("observers file '%s' already exists", obsFile)
			}
			if fileExists(expFile) {
				return fmt.Errorf("experiments file '%s' already exists", expFile)
			}

			for _, f := range []string{parFile, obsFile, expFile} {
				err := os.MkdirAll(filepath.Dir(f), os.ModePerm)
				if err != nil {
					return err
				}
			}

			p := params.CustomParams{
				Parameters: params.Default(),
				Custom:     map[reflect.Type]any{},
			}
			js, err := p.ToJSON()
			if err != nil {
				return err
			}
			err = writeBytes(parFile, js)
			if err != nil {
				return err
			}

			o := util.ObserversDef{
				Parameters:      "out/parameters.csv",
				CsvSeparator:    ",",
				TimeSeriesPlots: []util.TimeSeriesPlotDef{},
				Tables:          []util.TableDef{},
			}
			err = writeJSON(obsFile, &o)
			if err != nil {
				return err
			}
			e := util.ExperimentJs{
				Seed: 1,
				Parameters: []experiment.ParameterVariation{
					{
						Parameter: "params.InitialStores.Honey",
						SequenceFloatRange: &experiment.SequenceFloatRange{
							Min:    10,
							Max:    100,
							Values: 10,
						},
					},
				},
			}
			err = writeJSON(expFile, &e)
			if err != nil {
				return err
			}

			fmt.Printf("Successfully initialized experiment template in '%s'\n", dir)

			return nil
		},
	}
	root.Flags().StringVarP(&dir, "directory", "d", ".", "Working directory")

	return root
}

func writeJSON(path string, value any) error {
	js, err := json.MarshalIndent(value, "", "    ")
	if err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if _, err = f.Write(js); err != nil {
		return err
	}
	return f.Close()
}

func writeBytes(path string, value []byte) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if _, err = f.Write(value); err != nil {
		return err
	}
	return f.Close()
}

func fileExists(name string) bool {
	_, err := os.Stat(name)
	if err == nil {
		return true
	}
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	return false
}
