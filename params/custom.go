package params

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"github.com/mlange-42/arche/ecs"
	"github.com/mlange-42/beecs-cli/registry"
	baseparams "github.com/mlange-42/beecs/params"
)

type entry struct {
	Bytes []byte
}

func (e *entry) UnmarshalJSON(jsonData []byte) error {
	e.Bytes = jsonData
	return nil
}

func (e entry) MarshalJSON() ([]byte, error) {
	return e.Bytes, nil
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

	pars := customParamsJs{
		Params: p.Params,
	}
	err = decoder.Decode(&pars)
	if err != nil {
		return err
	}

	p.Params = pars.Params
	if p.Custom == nil {
		p.Custom = map[reflect.Type]any{}
	}

	for tpName, entry := range pars.Custom {
		tp, ok := registry.GetResource(tpName)
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

func (p *CustomParams) ToJSON() ([]byte, error) {
	par := customParamsJs{
		Params: p.Params,
		Custom: map[string]entry{},
	}

	for k, v := range p.Custom {
		js, err := json.MarshalIndent(&v, "", "    ")
		if err != nil {
			return []byte{}, err
		}
		par.Custom[k.String()] = entry{Bytes: js}
	}

	js, err := json.MarshalIndent(&par, "", "    ")
	if err != nil {
		return []byte{}, err
	}
	return js, nil
}

func (p *CustomParams) Apply(world *ecs.World) {
	p.Params.Apply(world)

	for tp, res := range p.Custom {
		id := ecs.ResourceTypeID(world, tp)
		world.Resources().Add(id, res)
	}
}
