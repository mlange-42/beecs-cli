package params

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"github.com/mlange-42/arche/ecs"
	"github.com/mlange-42/beecs-cli/util"
	baseparams "github.com/mlange-42/beecs/params"
)

type entry struct {
	Bytes []byte
}

func (e *entry) UnmarshalJSON(jsonData []byte) error {
	e.Bytes = jsonData
	return nil
}

type CustomParams struct {
	Params baseparams.DefaultParams
	Custom map[reflect.Type]any
}

type customParamsJs struct {
	Params baseparams.DefaultParams
	Custom map[string]entry
}

func (p *CustomParams) FromJSON(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()

	pars := customParamsJs{}
	err = decoder.Decode(&pars)
	if err != nil {
		return err
	}

	p.Params = pars.Params

	for tpName, entry := range pars.Custom {
		tp, ok := util.GetResource(tpName)
		if !ok {
			return fmt.Errorf("resource type '%s' is not registered", tpName)
		}
		resourceVal := reflect.New(tp).Interface()

		decoder := json.NewDecoder(bytes.NewReader(entry.Bytes))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&resourceVal); err != nil {
			return err
		}

		p.Custom[tp] = resourceVal
	}
	return nil
}

func (p *CustomParams) Apply(world *ecs.World) {
	p.Params.Apply(world)

	for tp, res := range p.Custom {
		id := ecs.ResourceTypeID(world, tp)
		world.Resources().Add(id, res)
	}
}
