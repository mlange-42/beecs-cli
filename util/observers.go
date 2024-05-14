package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"github.com/mlange-42/arche-model/observer"
	"github.com/mlange-42/arche-model/reporter"
	"github.com/mlange-42/arche-pixel/plot"
	"github.com/mlange-42/arche-pixel/window"
)

type entry struct {
	Bytes []byte
}

func (e *entry) UnmarshalJSON(jsonData []byte) error {
	e.Bytes = jsonData
	return nil
}

type timeSeriesPlot struct {
	Labels         plot.Labels
	Title          string
	Observer       string
	Params         entry
	Columns        []string
	Bounds         window.Bounds
	DrawInterval   int
	UpdateInterval int
	MaxRows        int
}

type table struct {
	File           string
	Observer       string
	Params         entry
	UpdateInterval int
}

type ObserversDef struct {
	Parameters      string
	CsvSeparator    string
	TimeSeriesPlots []timeSeriesPlot
	Tables          []table
}

func (obs *ObserversDef) CreateObservers() (Observers, error) {
	tsPlots := []*window.Window{}
	for _, p := range obs.TimeSeriesPlots {
		tp, ok := GetObserver(p.Observer)
		if !ok {
			return Observers{}, fmt.Errorf("observer type '%s' is not registered", p.Observer)
		}
		observerVal := reflect.New(tp).Interface()
		if len(p.Params.Bytes) == 0 {
			p.Params.Bytes = []byte("{}")
		}

		decoder := json.NewDecoder(bytes.NewReader(p.Params.Bytes))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&observerVal); err != nil {
			return Observers{}, err
		}
		obsCast, ok := observerVal.(observer.Row)
		if !ok {
			return Observers{}, fmt.Errorf("type '%s' is not a Row observer", tp.String())
		}
		win := &window.Window{
			Title:        p.Title,
			Bounds:       p.Bounds,
			DrawInterval: p.DrawInterval,
		}
		win = win.With(&plot.TimeSeries{
			Observer:       obsCast,
			Columns:        p.Columns,
			UpdateInterval: p.UpdateInterval,
			Labels:         p.Labels,
			MaxRows:        p.MaxRows,
		})
		win = win.With(&plot.Controls{})

		tsPlots = append(tsPlots, win)
	}

	tables := []*reporter.Callback{}
	for _, t := range obs.Tables {
		tp, ok := GetObserver(t.Observer)
		if !ok {
			return Observers{}, fmt.Errorf("observer type '%s' is not registered", t.Observer)
		}
		observerVal := reflect.New(tp).Interface()
		if len(t.Params.Bytes) == 0 {
			t.Params.Bytes = []byte("{}")
		}
		decoder := json.NewDecoder(bytes.NewReader(t.Params.Bytes))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&observerVal); err != nil {
			return Observers{}, err
		}
		obsCast, ok := observerVal.(observer.Row)
		if !ok {
			return Observers{}, fmt.Errorf("type '%s' is not a Row observer", tp.String())
		}
		rep := &reporter.Callback{
			Observer:       obsCast,
			UpdateInterval: t.UpdateInterval,
			HeaderCallback: func(header []string) {},
			Callback:       func(step int, row []float64) {},
		}
		tables = append(tables, rep)
	}

	return Observers{
		TimeSeriesPlots: tsPlots,
		Tables:          tables,
	}, nil
}

type Observers struct {
	TimeSeriesPlots []*window.Window
	Tables          []*reporter.Callback
}

func ObserversDefFromJSON(path string) (ObserversDef, error) {
	file, err := os.Open(path)
	if err != nil {
		return ObserversDef{}, err
	}
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()

	var obs ObserversDef
	if err = decoder.Decode(&obs); err != nil {
		return obs, err
	}
	return obs, nil
}
