package util

import (
	"fmt"
	"log"
	"path"
	"reflect"

	amod "github.com/mlange-42/arche-model/model"
	"github.com/mlange-42/arche-pixel/window"
	"github.com/mlange-42/beecs/experiment"
	"github.com/mlange-42/beecs/model"
	"github.com/mlange-42/beecs/params"
)

func RunSequential(
	p params.Params,
	exp *experiment.Experiment,
	observers *ObserversDef,
	overwrite []experiment.ParameterValue,
	dir string,
	totalRuns int, tps float64,
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

	for j := 0; j < totalRuns; j++ {
		result, err := runModel(p, exp, observers, overwrite, m, j, false)
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
	overwrite []experiment.ParameterValue,
	dir string,
	totalRuns int, threads int, tps float64,
) error {
	// Channel for sending jobs to workers (buffered!).
	jobs := make(chan int, totalRuns)
	// Channel for retrieving results / done messages (buffered!).
	results := make(chan Tables, totalRuns)

	// Start the workers.
	for w := 0; w < threads; w++ {
		go worker(jobs, results, p, exp, observers, overwrite, tps)
	}

	// Send the jobs. Does not block due to buffered channel.
	for j := 0; j < totalRuns; j++ {
		jobs <- j
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

func worker(jobs <-chan int, results chan<- Tables, p params.Params, exp *experiment.Experiment, observers *ObserversDef, overwrite []experiment.ParameterValue, tps float64) {
	m := amod.New()
	m.FPS = 30
	m.TPS = tps

	// Process incoming jobs.
	for j := range jobs {
		// Run the model.
		res, err := runModel(p, exp, observers, overwrite, m, j, true)
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
	overwrite []experiment.ParameterValue,
	m *amod.Model,
	idx int, parallel bool,
) (Tables, error) {
	model.Default(p, m)
	err := exp.ApplyValues(idx, &m.World)
	if err != nil {
		return Tables{}, err
	}

	for _, par := range overwrite {
		if err = model.SetParameter(&m.World, par.Parameter, par.Value); err != nil {
			return Tables{}, err
		}
	}

	values := exp.Values(idx)

	obs, err := observers.CreateObservers()
	if err != nil {
		log.Fatal(err)
	}

	result := Tables{
		Index:   idx,
		Headers: make([][]string, len(obs.Tables)+1),
		Data:    make([][][]float64, len(obs.Tables)+1),
	}

	result.Headers[0] = []string{"Run"}
	result.Data[0] = [][]float64{{float64(idx)}}
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

	if !parallel {
		for _, p := range obs.TimeSeriesPlots {
			m.AddUISystem(p)
		}
	}

	if parallel {
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
