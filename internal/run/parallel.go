package run

import (
	"fmt"
	"log"
	"path"

	amod "github.com/mlange-42/arche-model/model"
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
	systems []amod.System,
	overwrite []experiment.ParameterValue,
	dir string,
	threads int, tps float64, rng *rand.Rand,
) error {
	totalRuns := exp.TotalRuns()
	// Channel for sending jobs to workers (buffered!).
	jobs := make(chan job, totalRuns)
	// Channel for retrieving results / done messages (buffered!).
	results := make(chan util.Tables, totalRuns)

	seeds := make([]int32, totalRuns)
	for i := range seeds {
		seeds[i] = rng.Int31()
	}

	// Start the workers.
	for w := 0; w < threads; w++ {
		go worker(jobs, results, p, exp, observers, systems, overwrite, tps)
	}

	// Send the jobs. Does not block due to buffered channel.
	for j := 0; j < totalRuns; j++ {
		jobs <- job{Index: j, Seed: seeds[j]}
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
	for j := 0; j < totalRuns; j++ {
		result := <-results
		err = writer.Write(&result)
		if err != nil {
			return err
		}
		fmt.Printf("Run %5d/%d\n", result.Index, totalRuns)
	}

	return writer.Close()
}

func worker(jobs <-chan job, results chan<- util.Tables,
	p params.Params, exp *experiment.Experiment, observers *util.ObserversDef,
	systems []amod.System, overwrite []experiment.ParameterValue, tps float64) {

	m := amod.New()
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
