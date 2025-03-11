package run

import (
	"fmt"
	"log"
	"path"

	"github.com/mlange-42/ark-tools/app"
	"github.com/mlange-42/beecs-cli/internal/util"
	"github.com/mlange-42/beecs/experiment"
	"github.com/mlange-42/beecs/params"
	"golang.org/x/exp/rand"
)

type job struct {
	Index int
	Seed  int32
}

func RunParallel(
	p params.Params,
	exp *experiment.Experiment,
	observers *util.ObserversDef,
	systems []app.System,
	overwrite []experiment.ParameterValue,
	dir string,
	threads int, tps float64, rng *rand.Rand,
	indices []int,
) error {
	maxRuns := exp.TotalRuns()
	totalRuns := maxRuns
	if len(indices) > 0 {
		totalRuns = len(indices)
	}

	// Channel for sending jobs to workers (buffered!).
	jobs := make(chan job, totalRuns)
	// Channel for retrieving results / done messages (buffered!).
	results := make(chan util.Tables, totalRuns)

	seeds := make([]int32, maxRuns)
	for i := range seeds {
		seeds[i] = rng.Int31()
	}

	// Start the workers.
	for w := 0; w < threads; w++ {
		go worker(jobs, results, p, exp, observers, systems, overwrite, tps)
	}

	// Send the jobs. Does not block due to buffered channel.
	err := iterate(maxRuns, indices, func(idx int) error {
		jobs <- job{Index: idx, Seed: seeds[idx]}
		return nil
	})
	if err != nil {
		return err
	}
	close(jobs)

	paramsFile := observers.Parameters
	if len(paramsFile) == 0 {
		paramsFile = ""
	} else {
		paramsFile = path.Join(dir, paramsFile)
	}
	files := []string{paramsFile}
	for _, t := range observers.Tables {
		files = append(files, path.Join(dir, t.File))
	}
	for _, t := range observers.StepTables {
		files = append(files, path.Join(dir, t.File))
	}
	writer, err := util.NewCsvWriter(files, observers.CsvSeparator)
	if err != nil {
		return err
	}

	// Collect done messages.
	err = iterate(maxRuns, indices, func(idx int) error {
		result := <-results
		err = writer.Write(&result)
		if err != nil {
			return err
		}
		fmt.Printf("Run %5d/%d\n", result.Index, totalRuns)
		return nil
	})
	if err != nil {
		return err
	}

	return writer.Close()
}

func worker(jobs <-chan job, results chan<- util.Tables,
	p params.Params, exp *experiment.Experiment, observers *util.ObserversDef,
	systems []app.System, overwrite []experiment.ParameterValue, tps float64) {

	m := app.New()
	m.FPS = 30
	m.TPS = tps

	// Process incoming jobs.
	for j := range jobs {
		// Run the model.
		res, err := runModel(p, exp, observers, systems, overwrite, m, j.Index, j.Seed, true)
		if err != nil {
			log.Fatal(err)
		}
		// Send done message. Does not block due to buffered channel.
		results <- res
	}
}
