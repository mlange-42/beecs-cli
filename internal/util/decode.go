package util

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"github.com/mlange-42/arche-model/model"
	"github.com/mlange-42/beecs-cli/registry"
	"github.com/mlange-42/beecs/experiment"
	"github.com/mlange-42/beecs/params"
	"golang.org/x/exp/rand"
)

func ParametersFromFile(path string, params *params.DefaultParams) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	return decoder.Decode(params)
}

type experimentJs struct {
	Seed       uint32
	Parameters []experiment.ParameterVariation
}

func ExperimentFromFile(path string, runs int, seed int) (experiment.Experiment, *rand.Rand, error) {
	file, err := os.Open(path)
	if err != nil {
		return experiment.Experiment{}, nil, err
	}
	defer file.Close()

	var expJs experimentJs

	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&expJs); err != nil {
		return experiment.Experiment{}, nil, err
	}

	if seed == 0 {
		seed = int(expJs.Seed)
	} else if seed < 0 {
		seed = int(rand.Uint32())
	}
	rng := rand.New(rand.NewSource(uint64(seed)))

	exp, err := experiment.New(expJs.Parameters, rng, runs)
	if err != nil {
		return experiment.Experiment{}, nil, err
	}

	return exp, rng, nil
}

func ObserversDefFromFile(path string) (ObserversDef, error) {
	file, err := os.Open(path)
	if err != nil {
		return ObserversDef{}, err
	}
	defer file.Close()

	var obs ObserversDef

	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&obs); err != nil {
		return obs, err
	}

	if obs.CsvSeparator == "" {
		obs.CsvSeparator = ","
	}

	return obs, nil
}

func SystemsFromFile(path string) ([]model.System, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var sysStr []string

	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&sysStr); err != nil {
		return nil, err
	}

	sys := []model.System{}
	for _, tpName := range sysStr {
		tp, ok := registry.GetSystem(tpName)
		if !ok {
			return nil, fmt.Errorf("system type '%s' is not registered", tpName)
		}
		sysVal := reflect.New(tp).Interface()
		s, ok := sysVal.(model.System)
		if !ok {
			return nil, fmt.Errorf("system type '%s' does not implement the System interface", tpName)
		}
		sys = append(sys, s)
	}

	return sys, nil
}
