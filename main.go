package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	amod "github.com/mlange-42/arche-model/model"
	"github.com/mlange-42/arche-model/system"
	"github.com/mlange-42/arche-pixel/window"
	"github.com/mlange-42/beecs-cli/util"
	"github.com/mlange-42/beecs/experiment"
	"github.com/mlange-42/beecs/model"
	"github.com/mlange-42/beecs/params"
	"golang.org/x/exp/rand"
)

func main() {
	paramsFile := "_examples/base/parameters.json"
	expFile := "_examples/base/experiment.json"
	obsFile := "_examples/base/observers.json"

	p := params.Default()
	err := p.FromJSON(paramsFile)
	if err != nil {
		log.Fatal(err)
	}

	exp, err := ExperimentFromJSON(expFile)
	if err != nil {
		log.Fatal(err)
	}
	observers, err := util.ObserversDefFromJSON(obsFile)
	if err != nil {
		log.Fatal(err)
	}

	numSets := exp.ParameterSets()

	m := amod.New()

	for i := 0; i < numSets; i++ {
		model.Default(&p, m)
		exp.ApplyValues(i, &m.World)

		m.AddSystem(&system.FixedTermination{Steps: 365})

		obs, err := observers.CreateObservers()
		if err != nil {
			log.Fatal(err)
		}

		for _, t := range obs.Tables {
			m.AddSystem(t)
		}
		for _, p := range obs.TimeSeriesPlots {
			m.AddUISystem(p)
		}

		fmt.Printf("Run %d: %v\n", i, exp.Values(i))

		window.Run(m)
		//m.Run()
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
