package run

import (
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/mlange-42/ark-pixel/window"
	"github.com/mlange-42/ark-tools/app"
	"github.com/mlange-42/ark/ecs"
	"github.com/mlange-42/beecs-cli/internal/util"
	"github.com/mlange-42/beecs/experiment"
	"github.com/mlange-42/beecs/model"
	"github.com/mlange-42/beecs/params"
	butil "github.com/mlange-42/beecs/util"
)

func runModel(
	p params.Params,
	exp *experiment.Experiment,
	observers *util.ObserversDef,
	systems []app.System,
	overwrite []experiment.ParameterValue,
	a *app.App,
	idx int, rSeed int32, noUI bool,
) (util.Tables, error) {
	if len(systems) == 0 {
		model.Default(p, a)
	} else {
		sysCopy := make([]app.System, len(systems))
		// TODO: check copying!
		for i, sys := range systems {
			sysCopy[i] = butil.CopyInterface[app.System](sys)
		}
		model.WithSystems(p, sysCopy, a)
	}

	values := exp.Values(idx)
	err := exp.ApplyValues(values, &a.World)
	if err != nil {
		return util.Tables{}, err
	}

	for _, par := range overwrite {
		if err = model.SetParameter(&a.World, par.Parameter, par.Value); err != nil {
			return util.Tables{}, err
		}
	}

	seedRes := ecs.GetResource[params.RandomSeed](&a.World)
	if rSeed >= 0 && seedRes.Seed <= 0 {
		seedRes.Seed = int(rSeed)
		a.Seed(uint64(rSeed))
	}

	obs, err := observers.CreateObservers(!noUI)
	if err != nil {
		log.Fatal(err)
	}

	result := util.Tables{
		Index:   idx,
		Headers: make([][]string, len(obs.Tables)+len(obs.StepTables)+1),
		Data:    make([][][]float64, len(obs.Tables)+len(obs.StepTables)+1),
	}

	now := time.Now().UnixMilli()
	seed := ecs.GetResource[params.RandomSeed](&a.World).Seed
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
		a.AddSystem(t)
	}

	offset := len(obs.Tables)
	for i, t := range obs.StepTables {
		t.HeaderCallback = func(header []string) {
			h := make([]string, len(header)+2)
			h[0] = "Run"
			h[1] = "Ticks"
			copy(h[2:], header)
			result.Headers[offset+i+1] = h
		}
		t.Callback = func(step int, table [][]float64) {
			for _, row := range table {
				data := make([]float64, len(row)+2)
				data[0] = float64(idx)
				data[1] = float64(step)
				copy(data[2:], row)

				result.Data[offset+i+1] = append(result.Data[offset+i+1], data)
			}
		}
		a.AddSystem(t)
	}

	if !noUI {
		for _, p := range obs.Windows {
			a.AddUISystem(p)
		}
	}

	if noUI {
		a.Run()
	} else {
		window.Run(a)
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
