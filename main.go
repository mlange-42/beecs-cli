package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/mlange-42/beecs-cli/params"
	"github.com/mlange-42/beecs-cli/util"
	"github.com/mlange-42/beecs/experiment"
	baseparams "github.com/mlange-42/beecs/params"
	"github.com/spf13/cobra"
	"golang.org/x/exp/rand"
)

const (
	PARAMETERS = "parameters.json"
	OBSERVERS  = "observers.json"
	EXPERIMENT = "experiment.json"
)

func main() {
	if err := RootCommand().Execute(); err != nil {
		fmt.Printf("ERROR: %s\n", err.Error())
		fmt.Print("\nRun `beecs-cli -h` for help!\n\n")
		os.Exit(1)
	}
}

// RootCommand sets up the CLI
func RootCommand() *cobra.Command {
	var dir string
	var outDir string
	var paramFiles []string
	var expFile string
	var obsFile string
	var speed float64
	var threads int
	var runs int
	var overwrite []string
	var seed int

	root := &cobra.Command{
		Use:           "beecs-cli",
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

			rand.Seed(uint64(time.Now().UTC().Nanosecond()))

			if outDir == "" {
				outDir = dir
			}

			p := params.CustomParams{
				Params: baseparams.Default(),
			}
			for _, f := range paramFiles {
				err := p.FromJSON(path.Join(dir, f))
				if err != nil {
					return err
				}
			}
			if p.Params.InitialPatches.File != "" {
				p.Params.InitialPatches.File = path.Join(dir, p.Params.InitialPatches.File)
			}
			if !p.Params.ForagingPeriod.Builtin {
				for i, f := range p.Params.ForagingPeriod.Files {
					p.Params.ForagingPeriod.Files[i] = path.Join(dir, f)
				}
			}

			var exp experiment.Experiment
			var err error
			if expFile != "" {
				exp, err = util.ExperimentFromFile(path.Join(dir, expFile))
				if err != nil {
					return err
				}
				if seed > 0 {
					exp.Seed(uint64(seed))
				} else {
					exp.Seed(rand.Uint64())
				}
			}

			var observers util.ObserversDef
			if obsFile != "" {
				observers, err = util.ObserversDefFromFile(path.Join(dir, obsFile))
				if err != nil {
					return err
				}
			}

			numSets := exp.ParameterSets()
			if numSets == 0 {
				numSets = 1
			}
			totalRuns := numSets * runs
			if totalRuns == 1 {
				threads = 1
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
			if threads <= 1 {
				return util.RunSequential(&p, &exp, &observers, overwriteParams, outDir, totalRuns, speed, seed)
			} else {
				return util.RunParallel(&p, &exp, &observers, overwriteParams, outDir, totalRuns, threads, speed, seed)
			}
		},
	}

	root.Flags().StringVarP(&dir, "directory", "d", ".", "Working directory")
	root.Flags().StringVarP(&outDir, "output", "", "", "Output directory if different from working directory")
	root.Flags().StringSliceVarP(&paramFiles, "parameters", "p", []string{PARAMETERS}, "Parameter files, processed in the given order")
	root.Flags().StringVarP(&expFile, "experiment", "e", "", "Experiment file for parameter variation")
	root.Flags().StringVarP(&obsFile, "observers", "o", OBSERVERS, "Observers file")
	root.Flags().Float64VarP(&speed, "speed", "s", 0, "Speed limit in ticks per second. Default: 0 (unlimited)")
	root.Flags().IntVarP(&threads, "threads", "t", runtime.NumCPU(), "Number of threads")
	root.Flags().IntVarP(&runs, "runs", "r", 1, "Runs per parameter set")
	root.Flags().IntVarP(&seed, "seed", "", 0, "Super random seed for seed generation. Default: 0 (unseeded)")
	root.Flags().StringSliceVarP(&overwrite, "overwrite", "x", []string{}, "Overwrite variables like key1=value1,key2=value2")

	root.AddCommand(InitCommand())
	root.AddCommand(ParametersCommand())

	return root
}

func ParametersCommand() *cobra.Command {
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
			p := baseparams.Default()
			for _, f := range paramFiles {
				err := util.ParametersFromFile(path.Join(dir, f), &p)
				if err != nil {
					return err
				}
			}

			js, err := json.MarshalIndent(&p, "", "    ")
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

func InitCommand() *cobra.Command {
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
			parFile := path.Join(dir, PARAMETERS)
			obsFile := path.Join(dir, OBSERVERS)
			expFile := path.Join(dir, EXPERIMENT)

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

			p := baseparams.Default()
			err := writeJSON(parFile, &p)
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

			e := []experiment.ParameterVariation{
				{
					Parameter: "params.InitialStores.Honey",
					SequenceFloatRange: &experiment.SequenceFloatRange{
						Min:    10,
						Max:    100,
						Values: 10,
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
