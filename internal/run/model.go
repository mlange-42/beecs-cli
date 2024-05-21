package run

import (
	"fmt"
	"log"
	"reflect"
	"time"

	amod "github.com/mlange-42/arche-model/model"
	"github.com/mlange-42/arche-pixel/window"
	"github.com/mlange-42/arche/ecs"
	"github.com/mlange-42/beecs-cli/internal/util"
	"github.com/mlange-42/beecs/experiment"
	"github.com/mlange-42/beecs/model"
	"github.com/mlange-42/beecs/params"
)

func runModel(
	p params.Params,
	exp *experiment.Experiment,
	observers *util.ObserversDef,
	systems []amod.System,
	overwrite []experiment.ParameterValue,
	m *amod.Model,
	idx int, rSeed int32, noUi bool,
) (util.Tables, error) {
	if len(systems) == 0 {
		model.Default(p, m)
	} else {
		sysCopy := make([]amod.System, len(systems))
		// TODO: check copying!
		for i, sys := range systems {
			sysCopy[i] = util.CopyInterface[amod.System](sys)
		}
		model.WithSystems(p, sysCopy, m)
	}

	values := exp.Values(idx)
	err := exp.ApplyValues(values, &m.World)
	if err != nil {
		return util.Tables{}, err
	}

	for _, par := range overwrite {
		if err = model.SetParameter(&m.World, par.Parameter, par.Value); err != nil {
			return util.Tables{}, err
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

	result := util.Tables{
		Index:   idx,
		Headers: make([][]string, len(obs.Tables)+1),
		Data:    make([][][]float64, len(obs.Tables)+1),
	}

	now := time.Now().UnixMilli()
	seed := ecs.GetResource[params.RandomSeed](&m.World).Seed
	result.Headers[0] = []string{"Run", "Seed", "Started", "Finished"}
	result.Data[0] = [][]float64{{float64(idx), float64(seed), float64(now), 0}}
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

	now = time.Now().UnixMilli()
	result.Data[0][0][3] = float64(now)

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
