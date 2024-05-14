package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/mlange-42/beecs-cli/util"
	"github.com/mlange-42/beecs/experiment"
	"github.com/mlange-42/beecs/params"
	"golang.org/x/exp/rand"
)

func main() {
	paramsFile := "_examples/base/parameters.json"
	expFile := "_examples/base/experiment.json"
	obsFile := "_examples/base/observers.json"

	threads := 6

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
	runs := numSets * 10

	if threads <= 1 {
		util.RunSequential(&p, &exp, &observers, runs)
	} else {
		util.RunParallel(&p, &exp, &observers, runs, threads)
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
