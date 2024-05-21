package run

import (
	"fmt"
	"path"

	amod "github.com/mlange-42/arche-model/model"
	"github.com/mlange-42/beecs-cli/internal/util"
	"github.com/mlange-42/beecs/experiment"
	"github.com/mlange-42/beecs/params"
	"golang.org/x/exp/rand"
)

func RunSequential(
	p params.Params,
	exp *experiment.Experiment,
	observers *util.ObserversDef,
	systems []amod.System,
	overwrite []experiment.ParameterValue,
	dir string,
	tps float64, rng *rand.Rand,
) error {
	m := amod.New()
	m.FPS = 30
	m.TPS = tps

	paramsFile := observers.Parameters
	if len(paramsFile) == 0 {
		paramsFile = ""
	} else {
		paramsFile = path.Join(dir, paramsFile)
	}
	files := []string{paramsFile}
	for _, t := range observers.Tables {
		files = append(files, path.Join(dir, t.File))
	}
	writer, err := util.NewCsvWriter(files, observers.CsvSeparator)
	if err != nil {
		return err
	}

	totalRuns := exp.TotalRuns()
	for j := 0; j < totalRuns; j++ {
		result, err := runModel(p, exp, observers, systems, overwrite, m, j, rng.Int31(), totalRuns > 1)
		if err != nil {
			return err
		}
		err = writer.Write(&result)
		if err != nil {
			return err
		}
		fmt.Printf("Run %5d/%d\n", j, totalRuns)
	}

	return writer.Close()
}
