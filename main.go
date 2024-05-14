package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	amod "github.com/mlange-42/arche-model/model"
	"github.com/mlange-42/arche-model/system"
	"github.com/mlange-42/arche-pixel/window"
	"github.com/mlange-42/beecs-cli/util"
	"github.com/mlange-42/beecs/experiment"
	"github.com/mlange-42/beecs/model"
	"github.com/mlange-42/beecs/params"
	"golang.org/x/exp/rand"
)

func main() {
	paramsFile := "_examples/base/parameters.json"
	expFile := "_examples/base/experiment.json"
	obsFile := "_examples/base/observers.json"

	threads := 6

	p := params.Default()
	err := p.FromJSON(paramsFile)
	if err != nil {
		log.Fatal(err)
	}

	exp, err := ExperimentFromJSON(expFile)
	if err != nil {
		log.Fatal(err)
	}
	observers, err := util.ObserversDefFromJSON(obsFile)
	if err != nil {
		log.Fatal(err)
	}

	numSets := exp.ParameterSets()

	if threads <= 1 {
		runSequential(&p, &exp, &observers, numSets)
	} else {
		runParallel(&p, &exp, &observers, numSets, threads)
	}
}

func runSequential(p params.Params, exp *experiment.Experiment, observers *util.ObserversDef, totalRuns int) {
	m := amod.New()

	files := []string{}
	for _, t := range observers.Tables {
		files = append(files, t.File)
	}
	writer, err := util.NewCsvWriter(files)
	if err != nil {
		log.Fatal(err)
	}

	for j := 0; j < totalRuns; j++ {
		result := runModel(p, exp, observers, m, j, false)
		err = writer.Write(&result)
		if err != nil {
			log.Fatal(err)
		}
	}

	err = writer.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func runParallel(p params.Params, exp *experiment.Experiment, observers *util.ObserversDef, totalRuns int, threads int) {
	// Channel for sending jobs to workers (buffered!).
	jobs := make(chan int, totalRuns)
	// Channel for retrieving results / done messages (buffered!).
	results := make(chan util.Tables, totalRuns)

	// Start the workers.
	for w := 0; w < threads; w++ {
		go worker(jobs, results, p, exp, observers)
	}

	// Send the jobs. Does not block due to buffered channel.
	for j := 0; j < totalRuns; j++ {
		jobs <- j
	}
	close(jobs)

	files := []string{}
	for _, t := range observers.Tables {
		files = append(files, t.File)
	}
	writer, err := util.NewCsvWriter(files)
	if err != nil {
		log.Fatal(err)
	}

	// Collect done messages.
	for j := 0; j < totalRuns; j++ {
		result := <-results
		err = writer.Write(&result)
		if err != nil {
			log.Fatal(err)
		}
	}

	err = writer.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func worker(jobs <-chan int, results chan<- util.Tables, p params.Params, exp *experiment.Experiment, observers *util.ObserversDef) {
	m := amod.New()

	// Process incoming jobs.
	for j := range jobs {
		// Run the model.
		res := runModel(p, exp, observers, m, j, true)
		// Send done message. Does not block due to buffered channel.
		results <- res
	}
}

func runModel(p params.Params, exp *experiment.Experiment, observers *util.ObserversDef, m *amod.Model, idx int, parallel bool) util.Tables {
	model.Default(p, m)
	exp.ApplyValues(idx, &m.World)

	m.AddSystem(&system.FixedTermination{Steps: 365})

	obs, err := observers.CreateObservers()
	if err != nil {
		log.Fatal(err)
	}

	result := util.Tables{
		Index:   idx,
		Headers: make([][]string, len(obs.Tables)),
		Data:    make([][][]float64, len(obs.Tables)),
	}

	for i, t := range obs.Tables {
		t.HeaderCallback = func(header []string) {
			h := make([]string, len(header)+2)
			h[0] = "Run"
			h[1] = "Ticks"
			copy(h[2:], header)
			result.Headers[i] = h
		}
		t.Callback = func(step int, row []float64) {
			data := make([]float64, len(row)+2)
			data[0] = float64(idx)
			data[1] = float64(step)
			copy(data[2:], row)

			result.Data[i] = append(result.Data[i], data)
		}
		m.AddSystem(t)
	}

	if !parallel {
		for _, p := range obs.TimeSeriesPlots {
			m.AddUISystem(p)
		}
	}

	fmt.Printf("Run %d: %v\n", idx, exp.Values(idx))

	if parallel {
		m.Run()
	} else {
		window.Run(m)
	}

	return result
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
