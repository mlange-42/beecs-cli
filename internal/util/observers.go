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
)

type entry struct {
	Bytes []byte
}

func (e *entry) UnmarshalJSON(jsonData []byte) error {
	e.Bytes = jsonData
	return nil
}

type Observers struct {
	Windows    []*window.Window
	Tables     []*reporter.RowCallback
	StepTables []*reporter.TableCallback
}

type TimeSeriesPlotDef struct {
	Labels         plot.Labels
	Title          string
	Observer       string
	ObserverConfig entry
	Columns        []string
	Bounds         window.Bounds
	DrawInterval   int
	UpdateInterval int
	MaxRows        int
}

type LinePlotDef struct {
	Labels         plot.Labels
	Title          string
	Observer       string
	ObserverConfig entry
	X              string
	Y              []string
	Bounds         window.Bounds
	DrawInterval   int
	XLim           [2]float64
	YLim           [2]float64
}

type TableDef struct {
	File           string
	Observer       string
	ObserverConfig entry
	UpdateInterval int
	Final          bool
}

type StepTableDef struct {
	File           string
	Observer       string
	ObserverConfig entry
	UpdateInterval int
	Final          bool
}

type ViewDef struct {
	Drawer       string
	DrawerConfig entry
	Title        string
	Bounds       window.Bounds
	DrawInterval int
	MaxRows      int
}

type ObserversDef struct {
	Parameters      string              // Output file for parameters.
	CsvSeparator    string              // Column separator for all CSV output.
	TimeSeriesPlots []TimeSeriesPlotDef // Live time series plots.
	LinePlots       []LinePlotDef       // Live line plots.
	Views           []ViewDef           // Live views.
	Tables          []TableDef          // CSV output with one row per update.
	StepTables      []StepTableDef      // CSV output with a full table per update.
}

func (obs *ObserversDef) CreateObservers(withUI bool) (Observers, error) {
	windows := []*window.Window{}
	if withUI {
		win, err := createTimeSeriesPlots(obs.TimeSeriesPlots)
		if err != nil {
			return Observers{}, err
		}
		windows = append(windows, win...)

		win, err = createLinePlots(obs.LinePlots)
		if err != nil {
			return Observers{}, err
		}
		windows = append(windows, win...)

		win, err = createViews(obs.Views)
		if err != nil {
			return Observers{}, err
		}
		windows = append(windows, win...)
	}

	tables, err := createTables(obs.Tables)
	if err != nil {
		return Observers{}, err
	}

	stepTables, err := createStepTables(obs.StepTables)
	if err != nil {
		return Observers{}, err
	}

	return Observers{
		Windows:    windows,
		Tables:     tables,
		StepTables: stepTables,
	}, nil
}

func createTimeSeriesPlots(plots []TimeSeriesPlotDef) ([]*window.Window, error) {
	windows := make([]*window.Window, len(plots))
	for i, p := range plots {
		tp, ok := registry.GetObserver(p.Observer)
		if !ok {
			return nil, fmt.Errorf("observer type '%s' is not registered", p.Observer)
		}
		observerVal := reflect.New(tp).Interface()
		if len(p.ObserverConfig.Bytes) == 0 {
			p.ObserverConfig.Bytes = []byte("{}")
		}

		decoder := json.NewDecoder(bytes.NewReader(p.ObserverConfig.Bytes))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&observerVal); err != nil {
			return nil, err
		}
		obsCast, ok := observerVal.(observer.Row)
		if !ok {
			return nil, fmt.Errorf("type '%s' is not a Row observer", tp.String())
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

		windows[i] = win
	}

	return windows, nil
}

func createLinePlots(plots []LinePlotDef) ([]*window.Window, error) {
	windows := make([]*window.Window, len(plots))
	for i, p := range plots {
		tp, ok := registry.GetObserver(p.Observer)
		if !ok {
			return nil, fmt.Errorf("observer type '%s' is not registered", p.Observer)
		}
		observerVal := reflect.New(tp).Interface()
		if len(p.ObserverConfig.Bytes) == 0 {
			p.ObserverConfig.Bytes = []byte("{}")
		}

		decoder := json.NewDecoder(bytes.NewReader(p.ObserverConfig.Bytes))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&observerVal); err != nil {
			return nil, err
		}
		obsCast, ok := observerVal.(observer.Table)
		if !ok {
			return nil, fmt.Errorf("type '%s' is not a Table observer", tp.String())
		}
		win := &window.Window{
			Title:        p.Title,
			Bounds:       p.Bounds,
			DrawInterval: p.DrawInterval,
		}
		win = win.With(&plot.Lines{
			Observer: obsCast,
			X:        p.X,
			Y:        p.Y,
			Labels:   p.Labels,
			XLim:     p.XLim,
			YLim:     p.YLim,
		})
		win = win.With(&plot.Controls{})

		windows[i] = win
	}

	return windows, nil
}

func createViews(views []ViewDef) ([]*window.Window, error) {
	windows := make([]*window.Window, len(views))
	for i, p := range views {
		tp, ok := registry.GetDrawer(p.Drawer)
		if !ok {
			return nil, fmt.Errorf("view type '%s' is not registered", p.Drawer)
		}
		drawerVal := reflect.New(tp).Interface()
		if len(p.DrawerConfig.Bytes) == 0 {
			p.DrawerConfig.Bytes = []byte("{}")
		}

		decoder := json.NewDecoder(bytes.NewReader(p.DrawerConfig.Bytes))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&drawerVal); err != nil {
			return nil, err
		}
		drawerCast, ok := drawerVal.(window.Drawer)
		if !ok {
			return nil, fmt.Errorf("type '%s' is not a Drawer", tp.String())
		}
		win := &window.Window{
			Title:        p.Title,
			Bounds:       p.Bounds,
			DrawInterval: p.DrawInterval,
		}
		win = win.With(drawerCast)
		win = win.With(&plot.Controls{})

		windows[i] = win
	}
	return windows, nil
}

func createTables(tabs []TableDef) ([]*reporter.RowCallback, error) {
	tables := []*reporter.RowCallback{}
	for _, t := range tabs {
		tp, ok := registry.GetObserver(t.Observer)
		if !ok {
			return nil, fmt.Errorf("observer type '%s' is not registered", t.Observer)
		}
		observerVal := reflect.New(tp).Interface()
		if len(t.ObserverConfig.Bytes) == 0 {
			t.ObserverConfig.Bytes = []byte("{}")
		}
		decoder := json.NewDecoder(bytes.NewReader(t.ObserverConfig.Bytes))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&observerVal); err != nil {
			return nil, err
		}
		obsCast, ok := observerVal.(observer.Row)
		if !ok {
			return nil, fmt.Errorf("type '%s' is not a Row observer", tp.String())
		}
		rep := &reporter.RowCallback{
			Observer:       obsCast,
			UpdateInterval: t.UpdateInterval,
			HeaderCallback: func(header []string) {},
			Callback:       func(step int, row []float64) {},
			Final:          t.Final,
		}
		tables = append(tables, rep)
	}

	return tables, nil
}

func createStepTables(tabs []StepTableDef) ([]*reporter.TableCallback, error) {
	tables := []*reporter.TableCallback{}
	for _, t := range tabs {
		tp, ok := registry.GetObserver(t.Observer)
		if !ok {
			return nil, fmt.Errorf("observer type '%s' is not registered", t.Observer)
		}
		observerVal := reflect.New(tp).Interface()
		if len(t.ObserverConfig.Bytes) == 0 {
			t.ObserverConfig.Bytes = []byte("{}")
		}
		decoder := json.NewDecoder(bytes.NewReader(t.ObserverConfig.Bytes))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&observerVal); err != nil {
			return nil, err
		}
		obsCast, ok := observerVal.(observer.Table)
		if !ok {
			return nil, fmt.Errorf("type '%s' is not a Table observer", tp.String())
		}
		rep := &reporter.TableCallback{
			Observer:       obsCast,
			UpdateInterval: t.UpdateInterval,
			HeaderCallback: func(header []string) {},
			Callback:       func(step int, row [][]float64) {},
			Final:          t.Final,
		}
		tables = append(tables, rep)
	}

	return tables, nil
}
