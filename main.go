package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/mlange-42/beecs-cli/util"
	"github.com/mlange-42/beecs/experiment"
	"github.com/mlange-42/beecs/params"
	"github.com/spf13/cobra"
	"golang.org/x/exp/rand"
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

			p := params.Default()
			for _, f := range paramFiles {
				err := util.ParametersFromFile(path.Join(dir, f), &p)
				if err != nil {
					return err
				}
			}

			var exp experiment.Experiment
			var err error
			if expFile != "" {
				exp, err = util.ExperimentFromFile(path.Join(dir, expFile))
				if err != nil {
					return err
				}
			}
			if seed > 0 {
				exp.Seed(uint64(seed))
			} else {
				exp.Seed(rand.Uint64())
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
				return util.RunSequential(&p, &exp, &observers, overwriteParams, dir, totalRuns, speed, seed)
			} else {
				return util.RunParallel(&p, &exp, &observers, overwriteParams, dir, totalRuns, threads, speed, seed)
			}
		},
	}

	root.Flags().StringVarP(&dir, "directory", "d", ".", "Working directory")
	root.Flags().StringSliceVarP(&paramFiles, "parameters", "p", []string{"parameters.json"}, "Parameter files, processed in the given order")
	root.Flags().StringVarP(&expFile, "experiment", "e", "", "Experiment file for parameter variation")
	root.Flags().StringVarP(&obsFile, "observers", "o", "observers.json", "Observers file")
	root.Flags().Float64VarP(&speed, "speed", "s", 0, "Speed limit in ticks per second. Default: 0 (unlimited)")
	root.Flags().IntVarP(&threads, "threads", "t", runtime.NumCPU(), "Number of threads")
	root.Flags().IntVarP(&runs, "runs", "r", 1, "Runs per parameter set")
	root.Flags().IntVarP(&seed, "seed", "", 0, "Super random seed for seed generation. Default: 0 (unseeded)")
	root.Flags().StringSliceVarP(&overwrite, "overwrite", "x", []string{}, "Overwrite variables like key1=value1,key2=value2")

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
			p := params.Default()
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
