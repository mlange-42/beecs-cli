package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/mlange-42/arche-model/observer"
	"github.com/mlange-42/arche-model/reporter"
	"github.com/mlange-42/arche-pixel/plot"
	"github.com/mlange-42/arche-pixel/window"
	"github.com/mlange-42/beecs-cli/registry"
	"github.com/mlange-42/beecs-cli/view"
)

type entry struct {
	Bytes []byte
}

func (e *entry) UnmarshalJSON(jsonData []byte) error {
	e.Bytes = jsonData
	return nil
}

type TimeSeriesPlotDef struct {
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

type TableDef struct {
	File           string
	Observer       string
	Params         entry
	UpdateInterval int
	Final          bool
}

type ObserversDef struct {
	Parameters      string
	CsvSeparator    string
	TimeSeriesPlots []TimeSeriesPlotDef
	Tables          []TableDef
	Monitor         bool // Show the ECS monitor.
	Resources       bool // Show the resources inspector.
	Systems         bool // Show the systems inspector.
	ForagingView    bool // Show the flower patch foraging view.
}

func (obs *ObserversDef) CreateObservers(withUI bool) (Observers, error) {
	tsPlots := []*window.Window{}
	if withUI {
		for _, p := range obs.TimeSeriesPlots {
			tp, ok := registry.GetObserver(p.Observer)
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

		if obs.Monitor {
			win := (&window.Window{}).
				With(&plot.Monitor{}).
				With(&plot.Controls{})
			tsPlots = append(tsPlots, win)
		}

		if obs.Resources {
			win := (&window.Window{}).
				With(&plot.Resources{}).
				With(&plot.Controls{})
			tsPlots = append(tsPlots, win)
		}

		if obs.Systems {
			win := (&window.Window{}).
				With(&plot.Systems{}).
				With(&plot.Controls{})
			tsPlots = append(tsPlots, win)
		}

		if obs.ForagingView {
			win := (&window.Window{}).
				With(&view.Foraging{}).
				With(&plot.Controls{})
			tsPlots = append(tsPlots, win)
		}
	}

	tables := []*reporter.Callback{}
	for _, t := range obs.Tables {
		tp, ok := registry.GetObserver(t.Observer)
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
			Final:          t.Final,
		}
		tables = append(tables, rep)
	}

	return Observers{
		Windows: tsPlots,
		Tables:  tables,
	}, nil
}

type Observers struct {
	Windows []*window.Window
	Tables  []*reporter.Callback
}
