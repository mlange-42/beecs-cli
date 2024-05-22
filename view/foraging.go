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
)

type Foraging struct {
	drawer imdraw.IMDraw

	patchFilter generic.Filter2[comp.Coords, comp.Resource]
}

// Initialize the system
func (f *Foraging) Initialize(w *ecs.World, win *opengl.Window) {
	f.drawer = *imdraw.New(nil)

	f.patchFilter = *generic.NewFilter2[comp.Coords, comp.Resource]()
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

	drawCircle(dr, pixel.V(cx, cy), 5, 0, color.RGBA{128, 128, 128, 255})

	drawCircle(dr, pixel.V(cx, cy), 1000*scale, 1, color.RGBA{60, 60, 60, 255})
	drawCircle(dr, pixel.V(cx, cy), 2000*scale, 1, color.RGBA{60, 60, 60, 255})
	drawCircle(dr, pixel.V(cx, cy), 3000*scale, 1, color.RGBA{60, 60, 60, 255})
	drawCircle(dr, pixel.V(cx, cy), 4000*scale, 1, color.RGBA{60, 60, 60, 255})
	drawCircle(dr, pixel.V(cx, cy), 5000*scale, 1, color.RGBA{60, 60, 60, 255})

	query := f.patchFilter.Query(w)
	for query.Next() {
		coords, res := query.Get()
		px, py := cx+coords.X*scale, cy+coords.Y*scale

		nectar := res.Nectar * 0.000_001 * barHeight
		maxNectar := res.MaxNectar * 0.000_001 * barHeight
		pollen := res.Pollen * 0.001 * 20 * barHeight
		maxPollen := res.MaxPollen * 0.001 * 20 * barHeight

		drawCircle(dr, pixel.V(px, py), 3, 0, color.RGBA{128, 128, 128, 255})

		drawRect(dr, pixel.V(px-barWidth, py+nectar), pixel.V(px, py), 0, color.RGBA{180, 180, 0, 255})
		drawRect(dr, pixel.V(px, py+pollen), pixel.V(px+barWidth, py), 0, color.RGBA{180, 0, 180, 255})

		drawRect(dr, pixel.V(px-barWidth, py+maxNectar), pixel.V(px, py), 1, color.RGBA{180, 180, 80, 255})
		drawRect(dr, pixel.V(px, py+maxPollen), pixel.V(px+barWidth, py), 1, color.RGBA{180, 80, 180, 255})
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

func drawRect(dr *imdraw.IMDraw, p1, p2 pixel.Vec, thickness float64, color color.RGBA) {
	dr.Color = color
	dr.Push(p1, p2)
	dr.Rectangle(thickness)
	dr.Reset()
}
