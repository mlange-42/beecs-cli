package util

import (
	"fmt"
	"log"
	"path"
	"reflect"

	amod "github.com/mlange-42/arche-model/model"
	"github.com/mlange-42/arche-pixel/window"
	"github.com/mlange-42/arche/ecs"
	"github.com/mlange-42/beecs/experiment"
	"github.com/mlange-42/beecs/model"
	"github.com/mlange-42/beecs/params"
	"golang.org/x/exp/rand"
)

type job struct {
	Index int
	Seed  int32
}

func RunSequential(
	p params.Params,
	exp *experiment.Experiment,
	observers *ObserversDef,
	systems []amod.System,
	overwrite []experiment.ParameterValue,
	dir string,
	tps float64, seed int,
) error {
	m := amod.New()
	m.FPS = 30
	m.TPS = tps

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
	writer, err := NewCsvWriter(files, observers.CsvSeparator)
	if err != nil {
		return err
	}

	var rng *rand.Rand
	if seed > 0 {
		rng = rand.New(rand.NewSource(uint64(seed)))
	} else {
		rng = rand.New(rand.NewSource(rand.Uint64()))
	}
	totalRuns := exp.TotalRuns()
	if totalRuns == 0 {
		totalRuns = 1
	}
	for j := 0; j < totalRuns; j++ {
		result, err := runModel(p, exp, observers, systems, overwrite, m, j, rng.Int31(), totalRuns > 1)
		if err != nil {
			return err
		}
		err = writer.Write(&result)
		if err != nil {
			return err
		}
		fmt.Printf("Run %5d/%d\n", j, totalRuns)
	}

	return writer.Close()
}

func RunParallel(
	p params.Params,
	exp *experiment.Experiment,
	observers *ObserversDef,
	systems []amod.System,
	overwrite []experiment.ParameterValue,
	dir string,
	threads int, tps float64, seed int,
) error {
	totalRuns := exp.TotalRuns()
	// Channel for sending jobs to workers (buffered!).
	jobs := make(chan job, totalRuns)
	// Channel for retrieving results / done messages (buffered!).
	results := make(chan Tables, totalRuns)

	var rng *rand.Rand
	if seed > 0 {
		rng = rand.New(rand.NewSource(uint64(seed)))
	} else {
		rng = rand.New(rand.NewSource(uint64(rand.Int31())))
	}
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
	writer, err := NewCsvWriter(files, observers.CsvSeparator)
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

func worker(jobs <-chan job, results chan<- Tables,
	p params.Params, exp *experiment.Experiment, observers *ObserversDef,
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

func runModel(
	p params.Params,
	exp *experiment.Experiment,
	observers *ObserversDef,
	systems []amod.System,
	overwrite []experiment.ParameterValue,
	m *amod.Model,
	idx int, rSeed int32, noUi bool,
) (Tables, error) {
	if len(systems) == 0 {
		model.Default(p, m)
	} else {
		sysCopy := make([]amod.System, len(systems))
		// TODO: check copying!
		for i, sys := range systems {
			sysCopy[i] = CopyInterface[amod.System](sys)
		}
		model.WithSystems(p, sysCopy, m)
	}

	values := exp.Values(idx)
	err := exp.ApplyValues(values, &m.World)
	if err != nil {
		return Tables{}, err
	}

	for _, par := range overwrite {
		if err = model.SetParameter(&m.World, par.Parameter, par.Value); err != nil {
			return Tables{}, err
		}
	}
	if rSeed >= 0 {
		ecs.GetResource[params.RandomSeed](&m.World).Seed = int(rSeed)
		m.Seed(uint64(rSeed))
	}

	obs, err := observers.CreateObservers()
	if err != nil {
		log.Fatal(err)
	}

	result := Tables{
		Index:   idx,
		Headers: make([][]string, len(obs.Tables)+1),
		Data:    make([][][]float64, len(obs.Tables)+1),
	}

	seed := ecs.GetResource[params.RandomSeed](&m.World).Seed
	result.Headers[0] = []string{"Run", "Seed"}
	result.Data[0] = [][]float64{{float64(idx), float64(seed)}}
	for _, v := range values {
		result.Headers[0] = append(result.Headers[0], v.Parameter)
		floatValue := toFloat(v.Value)
		result.Data[0][0] = append(result.Data[0][0], floatValue)
	}

	for i, t := range obs.Tables {
		t.HeaderCallback = func(header []string) {
			h := make([]string, len(header)+2)
			h[0] = "Run"
			h[1] = "Ticks"
			copy(h[2:], header)
			result.Headers[i+1] = h
		}
		t.Callback = func(step int, row []float64) {
			data := make([]float64, len(row)+2)
			data[0] = float64(idx)
			data[1] = float64(step)
			copy(data[2:], row)

			result.Data[i+1] = append(result.Data[i+1], data)
		}
		m.AddSystem(t)
	}

	if !noUi {
		for _, p := range obs.TimeSeriesPlots {
			m.AddUISystem(p)
		}
	}

	if noUi {
		m.Run()
	} else {
		window.Run(m)
	}

	return result, nil
}

func toFloat(v any) float64 {
	var floatValue float64
	switch vv := v.(type) {
	case float64:
		floatValue = vv
	case float32:
		floatValue = float64(vv)
	case int:
		floatValue = float64(vv)
	case int32:
		floatValue = float64(vv)
	case int64:
		floatValue = float64(vv)
	case bool:
		if vv {
			floatValue = 1
		}
	default:
		panic(fmt.Sprintf("unsupported parameter type %s", reflect.TypeOf(vv).String()))
	}

	return floatValue
}
