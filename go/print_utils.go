package main

import "github.com/signintech/gopdf"

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
