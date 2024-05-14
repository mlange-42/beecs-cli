package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	amod "github.com/mlange-42/arche-model/model"
	"github.com/mlange-42/arche-model/system"
	"github.com/mlange-42/beecs/experiment"
	"github.com/mlange-42/beecs/model"
	"github.com/mlange-42/beecs/params"
	"golang.org/x/exp/rand"
)

func main() {
	paramsFile := "_examples/base/parameters.json"
	expFile := "_examples/base/experiment.json"

	p := params.Default()
	err := p.FromJSON(paramsFile)
	if err != nil {
		log.Fatal(err)
	}

	exp, err := ExperimentFromJSON(expFile)
	if err != nil {
		log.Fatal(err)
	}

	numSets := exp.ParameterSets()

	m := amod.New()

	for i := 0; i < numSets; i++ {
		model.Default(&p, m)
		exp.ApplyValues(i, &m.World)

		m.AddSystem(&system.FixedTermination{Steps: 365})

		fmt.Printf("Run %d: %v\n", i, exp.Values(i))
		m.Run()
	}
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
