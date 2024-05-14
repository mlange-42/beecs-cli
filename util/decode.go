package util

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mlange-42/beecs/experiment"
	"github.com/mlange-42/beecs/params"
	toml "github.com/pelletier/go-toml/v2"
	"golang.org/x/exp/rand"
)

func ParametersFromFile(path string, params *params.DefaultParams) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	if strings.HasSuffix(path, ".json") || strings.HasSuffix(path, ".JSON") {
		decoder := json.NewDecoder(file)
		decoder.DisallowUnknownFields()
		return decoder.Decode(params)
	} else if strings.HasSuffix(path, ".toml") || strings.HasSuffix(path, ".TOML") {
		decoder := toml.NewDecoder(file)
		decoder.DisallowUnknownFields()
		return decoder.Decode(params)
	} else {
		return fmt.Errorf("only JSON and TOML supported")
	}
}

type variations struct {
	Experiment []experiment.ParameterVariation
}

func ExperimentFromFile(path string) (experiment.Experiment, error) {
	file, err := os.Open(path)
	if err != nil {
		return experiment.Experiment{}, err
	}

	var exp variations

	if strings.HasSuffix(path, ".json") || strings.HasSuffix(path, ".JSON") {
		decoder := json.NewDecoder(file)
		decoder.DisallowUnknownFields()
		if err = decoder.Decode(&exp); err != nil {
			return experiment.Experiment{}, err
		}
	} else if strings.HasSuffix(path, ".toml") || strings.HasSuffix(path, ".TOML") {
		decoder := toml.NewDecoder(file)
		decoder.DisallowUnknownFields()
		if err = decoder.Decode(&exp); err != nil {
			return experiment.Experiment{}, err
		}
	} else {
		return experiment.Experiment{}, fmt.Errorf("only JSON and TOML supported")
	}

	return experiment.New(exp.Experiment, rand.New(rand.NewSource(uint64(time.Now().UnixNano()))))
}

func ObserversDefFromFile(path string) (ObserversDef, error) {
	file, err := os.Open(path)
	if err != nil {
		return ObserversDef{}, err
	}
	var obs ObserversDef

	if strings.HasSuffix(path, ".json") || strings.HasSuffix(path, ".JSON") {
		decoder := json.NewDecoder(file)
		decoder.DisallowUnknownFields()
		if err = decoder.Decode(&obs); err != nil {
			return obs, err
		}
	} else if strings.HasSuffix(path, ".toml") || strings.HasSuffix(path, ".TOML") {
		decoder := toml.NewDecoder(file)
		decoder.DisallowUnknownFields()
		decoder.EnableUnmarshalerInterface()
		if err = decoder.Decode(&obs); err != nil {
			return obs, err
		}
	} else {
		return obs, fmt.Errorf("only JSON and TOML supported")
	}

	return obs, nil
}
