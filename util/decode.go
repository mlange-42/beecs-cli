package util

import (
	"encoding/json"
	"os"

	"github.com/mlange-42/beecs/experiment"
	"github.com/mlange-42/beecs/params"
	"golang.org/x/exp/rand"
)

func ParametersFromFile(path string, params *params.DefaultParams) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	return decoder.Decode(params)
}

func ExperimentFromFile(path string) (experiment.Experiment, error) {
	file, err := os.Open(path)
	if err != nil {
		return experiment.Experiment{}, err
	}

	var exp []experiment.ParameterVariation

	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&exp); err != nil {
		return experiment.Experiment{}, err
	}

	return experiment.New(exp, rand.New(nil))
}

func ObserversDefFromFile(path string) (ObserversDef, error) {
	file, err := os.Open(path)
	if err != nil {
		return ObserversDef{}, err
	}
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
