package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/mlange-42/beecs-cli/util"
	"github.com/mlange-42/beecs/experiment"
	"github.com/mlange-42/beecs/params"
	"github.com/spf13/cobra"
	"golang.org/x/exp/rand"
)

func main() {
	if err := RootCommand().Execute(); err != nil {
		fmt.Printf("%s\n", err.Error())
		os.Exit(1)
	}
}

func ExperimentFromJSON(path string) (experiment.Experiment, error) {
	file, err := os.Open(path)
	if err != nil {
		return experiment.Experiment{}, err
	}
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()

	var exp []experiment.ParameterVariation
	if err = decoder.Decode(&exp); err != nil {
		return experiment.Experiment{}, err
	}
	return experiment.New(exp, rand.New(rand.NewSource(uint64(time.Now().UnixNano()))))
}

// RootCommand sets up the CLI
func RootCommand() *cobra.Command {
	var paramsFile string
	var expFile string
	var obsFile string
	var tps float64
	var threads int
	var runs int

	root := &cobra.Command{
		Use:           "beecs-cli",
		Short:         "beecs-cli provides a command line interface for the beecs model",
		Long:          `beecs-cli provides a command line interface for the beecs model`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if paramsFile == "" {
				_ = cmd.Help()
				os.Exit(0)
			}

			p := params.Default()
			err := p.FromJSON(paramsFile)
			if err != nil {
				return err
			}

			var exp experiment.Experiment
			if expFile != "" {
				exp, err = ExperimentFromJSON(expFile)
				if err != nil {
					return err
				}
			}
			var observers util.ObserversDef
			if obsFile != "" {
				observers, err = util.ObserversDefFromJSON(obsFile)
				if err != nil {
					return err
				}
			}

			numSets := exp.ParameterSets()
			if numSets == 0 {
				numSets = 1
			}
			totalRuns := numSets * runs

			if threads <= 1 {
				return util.RunSequential(&p, &exp, &observers, totalRuns, tps)
			} else {
				return util.RunParallel(&p, &exp, &observers, totalRuns, threads, tps)
			}
		},
	}
	root.Flags().StringVarP(&paramsFile, "parameters", "p", "", "Parameters file.")
	root.Flags().StringVarP(&expFile, "experiment", "e", "", "Experiment file.")
	root.Flags().StringVarP(&obsFile, "observers", "o", "", "Observers file.")
	root.Flags().Float64VarP(&tps, "tps", "s", 0, "Limit ticks per second.")
	root.Flags().IntVarP(&threads, "threads", "t", runtime.NumCPU(), "Number of threads.")
	root.Flags().IntVarP(&runs, "runs", "r", 1, "Runs per parameter set.")

	return root
}
