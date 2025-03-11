package run

import (
	"fmt"
	"path"

	"github.com/mlange-42/ark-tools/app"
	"github.com/mlange-42/beecs-cli/internal/util"
	"github.com/mlange-42/beecs/experiment"
	"github.com/mlange-42/beecs/params"
	"golang.org/x/exp/rand"
)

func RunSequential(
	p params.Params,
	exp *experiment.Experiment,
	observers *util.ObserversDef,
	systems []app.System,
	overwrite []experiment.ParameterValue,
	dir string,
	tps float64, rng *rand.Rand,
	indices []int,
) error {
	m := app.New()
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
	for _, t := range observers.StepTables {
		files = append(files, path.Join(dir, t.File))
	}
	writer, err := util.NewCsvWriter(files, observers.CsvSeparator)
	if err != nil {
		return err
	}

	maxRuns := exp.TotalRuns()
	actualRuns := maxRuns
	if len(indices) > 0 {
		actualRuns = len(indices)
	}
	err = iterate(maxRuns, indices, func(idx int) error {
		result, err := runModel(p, exp, observers, systems, overwrite, m, idx, rng.Int31(), actualRuns > 1)
		if err != nil {
			return err
		}
		err = writer.Write(&result)
		if err != nil {
			return err
		}
		fmt.Printf("Run %5d/%d\n", idx, maxRuns)
		return nil
	})
	if err != nil {
		return err
	}

	return writer.Close()
}

func iterate(totalRuns int, indices []int, fn func(idx int) error) error {
	if len(indices) == 0 {
		for j := 0; j < totalRuns; j++ {
			if err := fn(j); err != nil {
				return err
			}
		}
		return nil
	}
	for _, j := range indices {
		if err := fn(j); err != nil {
			return err
		}
	}
	return nil
}
