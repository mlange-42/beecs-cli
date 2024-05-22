package view

import (
	"image/color"
	"math"

	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"github.com/gopxl/pixel/v2/ext/imdraw"
	"github.com/mlange-42/arche/ecs"
	"github.com/mlange-42/arche/generic"
	"github.com/mlange-42/beecs/comp"
	"github.com/mlange-42/beecs/globals"
	"github.com/mlange-42/beecs/params"
)

type Foraging struct {
	drawer imdraw.IMDraw

	stores        *globals.Stores
	popStats      *globals.PopulationStats
	energyContent *params.EnergyContent
	patchFilter   generic.Filter3[comp.Coords, comp.Resource, comp.Visits]
}

// Initialize the system
func (f *Foraging) Initialize(w *ecs.World, win *opengl.Window) {
	f.drawer = *imdraw.New(nil)

	f.stores = ecs.GetResource[globals.Stores](w)
	f.popStats = ecs.GetResource[globals.PopulationStats](w)
	f.energyContent = ecs.GetResource[params.EnergyContent](w)
	f.patchFilter = *generic.NewFilter3[comp.Coords, comp.Resource, comp.Visits]()
}

// Update the drawer.
func (f *Foraging) Update(w *ecs.World) {}

// UpdateInputs handles input events of the previous frame update.
func (f *Foraging) UpdateInputs(w *ecs.World, win *opengl.Window) {}

// Draw the system
func (f *Foraging) Draw(w *ecs.World, win *opengl.Window) {
	width := win.Canvas().Bounds().W()
	height := win.Canvas().Bounds().H()

	dMax := 10_100.0

	scale := math.Min(width/dMax, height/dMax)

	cx := width / 2
	cy := height / 2
	barWidth := 8.0
	barHeight := 2.0

	dr := &f.drawer

	// Distance circles
	drawCircle(dr, pixel.V(cx, cy), 1000*scale, 1, color.RGBA{60, 60, 60, 255})
	drawCircle(dr, pixel.V(cx, cy), 2000*scale, 1, color.RGBA{60, 60, 60, 255})
	drawCircle(dr, pixel.V(cx, cy), 3000*scale, 1, color.RGBA{60, 60, 60, 255})
	drawCircle(dr, pixel.V(cx, cy), 4000*scale, 1, color.RGBA{60, 60, 60, 255})
	drawCircle(dr, pixel.V(cx, cy), 5000*scale, 1, color.RGBA{60, 60, 60, 255})

	// Hive resources
	honeyStore := f.stores.Honey / (1000.0 * f.energyContent.Honey)
	decentHoney := f.stores.DecentHoney / (1000.0 * f.energyContent.Honey)
	pollenStore := f.stores.Pollen * 0.001 * 20 * barHeight
	idealPollen := f.stores.IdealPollen * 0.001 * 20 * barHeight

	drawRect(dr, pixel.V(cx-barWidth, cy+honeyStore), pixel.V(cx, cy), 0, color.RGBA{180, 180, 0, 255})
	drawRect(dr, pixel.V(cx, cy+pollenStore), pixel.V(cx+barWidth, cy), 0, color.RGBA{180, 0, 180, 255})

	drawRect(dr, pixel.V(cx-barWidth, cy+decentHoney), pixel.V(cx, cy), 1, color.RGBA{180, 180, 120, 255})
	drawRect(dr, pixel.V(cx, cy+idealPollen), pixel.V(cx+barWidth, cy), 1, color.RGBA{180, 120, 180, 255})

	// Hive age classes
	popScale := 0.2
	popLine := 0.0
	drawArcS(dr, pixel.V(cx, cy),
		popScale*math.Sqrt(float64(f.popStats.TotalPopulation)),
		popLine, color.RGBA{128, 128, 128, 255})

	drawArcS(dr, pixel.V(cx, cy),
		popScale*math.Sqrt(float64(f.popStats.TotalBrood)),
		popLine, color.RGBA{230, 230, 230, 255})

	query := f.patchFilter.Query(w)
	for query.Next() {
		coords, res, vis := query.Get()
		px, py := cx+coords.X*scale, cy+coords.Y*scale

		// Patch marker
		drawCircle(dr, pixel.V(px, py), 3, 0, color.RGBA{128, 128, 128, 255})

		// Visits
		if vis.Nectar > 0 {
			drawArcSW(dr, pixel.V(px, py), math.Log2(float64(vis.Nectar)), 2, color.RGBA{180, 180, 80, 255})
		}
		if vis.Pollen > 0 {
			drawArcSE(dr, pixel.V(px, py), math.Log2(float64(vis.Pollen)), 2, color.RGBA{180, 80, 180, 255})
		}

		// Resource bars
		nectar := res.Nectar * 0.000_001 * barHeight
		maxNectar := res.MaxNectar * 0.000_001 * barHeight
		pollen := res.Pollen * 0.001 * 20 * barHeight
		maxPollen := res.MaxPollen * 0.001 * 20 * barHeight

		if maxNectar > 0 {
			drawRect(dr, pixel.V(px-barWidth, py+nectar), pixel.V(px, py), 0, color.RGBA{180, 180, 0, 255})
			drawRect(dr, pixel.V(px-barWidth, py+maxNectar), pixel.V(px, py), 1, color.RGBA{180, 180, 80, 255})
		}

		if maxPollen > 0 {
			drawRect(dr, pixel.V(px, py+pollen), pixel.V(px+barWidth, py), 0, color.RGBA{180, 0, 180, 255})
			drawRect(dr, pixel.V(px, py+maxPollen), pixel.V(px+barWidth, py), 1, color.RGBA{180, 80, 180, 255})
		}
	}

	dr.Draw(win)
	dr.Clear()
}

func drawCircle(dr *imdraw.IMDraw, center pixel.Vec, radius float64, thickness float64, color color.RGBA) {
	dr.Color = color
	dr.Push(center)
	dr.Circle(radius, thickness)
	dr.Reset()
}

func drawArcSW(dr *imdraw.IMDraw, center pixel.Vec, radius float64, thickness float64, color color.RGBA) {
	dr.Color = color
	dr.Push(center)
	dr.CircleArc(radius, math.Pi, math.Pi*1.5, thickness)
	dr.Reset()
}

func drawArcSE(dr *imdraw.IMDraw, center pixel.Vec, radius float64, thickness float64, color color.RGBA) {
	dr.Color = color
	dr.Push(center)
	dr.CircleArc(radius, math.Pi*1.5, math.Pi*2, thickness)
	dr.Reset()
}

func drawArcS(dr *imdraw.IMDraw, center pixel.Vec, radius float64, thickness float64, color color.RGBA) {
	dr.Color = color
	dr.Push(center)
	dr.CircleArc(radius, math.Pi, math.Pi*2, thickness)
	dr.Reset()
}

func drawRect(dr *imdraw.IMDraw, p1, p2 pixel.Vec, thickness float64, color color.RGBA) {
	dr.Color = color
	dr.Push(p1, p2)
	dr.Rectangle(thickness)
	dr.Reset()
}
