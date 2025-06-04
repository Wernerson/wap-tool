package main

import (
	"math"

	"github.com/signintech/gopdf"
)

// Draw a grid on pdf where start is the top left corner and bounds is the size of the grid.
// rows and columns define the number of rows and columns.
func drawGrid(pdf *gopdf.GoPdf, start gopdf.Point, bounds gopdf.Rect, rows, columns int) {
	drawHorizontalLines(pdf, start, bounds, rows)
	drawVerticalLines(pdf, start, bounds, columns)
}

func drawHorizontalLines(pdf *gopdf.GoPdf, start gopdf.Point, bounds gopdf.Rect, rows int) {
	rowHeight := bounds.H / float64(rows)
	// Draw horizontal lines ---
	for h := range rows + 1 {
		y := start.Y + rowHeight*float64(h)
		pdf.Line(start.X, y, start.X+bounds.W, y)
	}
}

func drawVerticalLines(pdf *gopdf.GoPdf, start gopdf.Point, bounds gopdf.Rect, columns int) {
	colWidth := bounds.W / float64(columns)
	// Draw vertical lines |
	for w := range columns + 1 {
		x := start.X + colWidth*float64(w)
		pdf.Line(x, start.Y, x, start.Y+bounds.H)
	}
}

// shorthand
func drawRect(pdf *gopdf.GoPdf, p gopdf.Point, rect gopdf.Rect) {
	err := pdf.Rectangle(p.X, p.Y, p.X+rect.W, p.Y+rect.H, "DF", 0, 0)
	checkPDFerror(err)
}

func drawOpenEndRect(pdf *gopdf.GoPdf, p gopdf.Point, rect gopdf.Rect) {
	offsetHeight := 5.0
	points := []gopdf.Point{}
	points = append(points, gopdf.Point{X: p.X, Y: p.Y})
	points = append(points, gopdf.Point{X: p.X, Y: p.Y + rect.H + offsetHeight})
	numSinPoints := 100
	for i := 0; i < numSinPoints; i++ {
		t := float64(i) / float64(numSinPoints)
		sinY := math.Sin(-2 * math.Pi * t)
		points = append(points, gopdf.Point{X: p.X + t*rect.W, Y: p.Y + rect.H + offsetHeight + sinY*offsetHeight})
	}
	points = append(points, gopdf.Point{X: p.X + rect.W, Y: p.Y + rect.H + offsetHeight})
	points = append(points, gopdf.Point{X: p.X + rect.W, Y: p.Y})
	pdf.Polygon(points, "DF")
}
