package main

import (
	"time"

	"github.com/signintech/gopdf"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func MakePDF(wap *Wap, outputPath string) (err error) {
	pdf := gopdf.GoPdf{}
	mm6ToPx := 22.68
	pageSize := gopdf.PageSizeA4Landscape
	trimbox := gopdf.Box{Left: mm6ToPx, Top: mm6ToPx, Right: pageSize.W - mm6ToPx, Bottom: pageSize.H - mm6ToPx}
	pdf.Start(gopdf.Config{
		PageSize: *pageSize,
		TrimBox:  trimbox,
	})
	pdf.AddPage()
	err = pdf.AddTTFFont("OpenSans", "./OpenSans-Regular.ttf")
	if err != nil {
		return err
	}
	err = pdf.SetFont("OpenSans", "", 14)
	if err != nil {
		return err
	}
	unit := ""
	if wap.data.Meta.Unit != nil {
		unit = *wap.data.Meta.Unit
	}
	version := time.Now().Format(time.DateOnly)
	if wap.data.Meta.Version != nil {
		version = *wap.data.Meta.Version
	}
	padding := mm6ToPx / 4
	pdf.AddHeader(func() {
		pdf.SetY(padding)
		pdf.SetX(padding)
		pdf.Cell(nil, unit)
		tm := wap.data.Meta.Title
		tmW, err := pdf.MeasureTextWidth(tm)
		check(err)
		pdf.SetX(pageSize.W/2 - tmW/2)
		pdf.CellWithOption(nil, tm, gopdf.CellOption{Align: gopdf.Center})
		tr := version
		trW, err := pdf.MeasureTextWidth(tr)
		check(err)
		pdf.SetX(pageSize.W - trW - padding)
		pdf.CellWithOption(nil, tr, gopdf.CellOption{Align: gopdf.Right})
	})
	pdf.AddFooter(func() {
		footerText := "footer"
		ftH, err := pdf.MeasureCellHeightByText(footerText)
		check(err)
		pdf.SetY(pageSize.H - padding - ftH)
		pdf.Cell(nil, "footer")
	})
	// Page trim-box
	opt := gopdf.PageOption{
		PageSize: *&pageSize,
		TrimBox:  &trimbox,
	}
	pdf.AddPageWithOption(opt)

	// [ ] Find the correct Start and bounds
	pdf.SetStrokeColor(255, 0, 0)
	pdf.SetLineWidth(1)
	Grid(&pdf, gopdf.Point{X: 50, Y: 50}, gopdf.Rect{W: 400, H: 400}, 5, 5)
	pdf.SetStrokeColor(0x80, 0x80, 0x80)
	pdf.SetLineWidth(0.5)
	Grid(&pdf, gopdf.Point{X: 50, Y: 50}, gopdf.Rect{W: 400, H: 400}, 20, 20)
	// [ ] Create based on Time settings
	// [ ] Convert into local coordinates
	// [ ] Add time labels
	// [ ] check for different page sizes

	// pdf.Cell(nil, "Hi")
	pdf.AddPage()
	pdf.SetY(400)
	pdf.Text("page 2 content")
	pdf.WritePdf(outputPath)
	return nil
}

// Draw a grid on pdf where start is the top left corner and bounds is the size of the grid.
// rows and columns define the number of rows and columns.
func Grid(pdf *gopdf.GoPdf, start gopdf.Point, bounds gopdf.Rect, rows, columns int) {
	rowHeight := bounds.H / float64(rows)
	colWidth := bounds.W / float64(columns)
	// Draw horizontal lines ---
	for h := range rows + 1 {
		y := start.Y + rowHeight*float64(h)
		pdf.Line(start.X, y, start.X+bounds.W, y)
	}
	// Draw vertical lines |
	for w := range columns + 1 {
		x := start.X + colWidth*float64(w)
		pdf.Line(x, start.Y, x, start.Y+bounds.H)
	}
	return nil
}
