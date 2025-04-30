package main

import (
	"fmt"
	"log"
	"time"

	"github.com/signintech/gopdf"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

// Convert milimeters to pixels
func mmToPx(mm float64) float64 {
	return 3.78 * mm
}

func MakePDF(wap *Wap, outputPath string) (err error) {
	pdf := gopdf.GoPdf{}
	mm6ToPx := mmToPx(6)
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
	padding := mmToPx(1.5)
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

	pdf.SetStrokeColor(255, 0, 0)
	pdf.SetLineWidth(1)
	PL := mmToPx(25)
	PR := PL
	PT := mmToPx(20)
	PB := mmToPx(30)
	P1 := gopdf.Point{X: PL, Y: PL}
	wapBox := gopdf.Rect{W: pageSize.W - PL - PR, H: pageSize.H - PT - PB}
	// [ ] Find the correct Start and bounds
	DAYS := 7
	duration := wap.dayEnd.Sub(wap.dayStart)
	HOURS := int(duration.Hours())
	SMALL_COLS := 5
	colWidth := wapBox.W / float64(DAYS)
	minuteHeight := wapBox.H / duration.Minutes()
	ToGridSystem := func(t time.Time, dayIndex int) gopdf.Point {
		deltaX := float64(dayIndex) * colWidth
		deltaY := t.Sub(wap.dayStart).Minutes() * minuteHeight
		return Add(P1, gopdf.Point{X: deltaX, Y: deltaY})
	}
	Grid(&pdf, P1, wapBox, HOURS, DAYS)
	pdf.SetStrokeColor(0x80, 0x80, 0x80)
	pdf.SetLineWidth(0.5)
	Grid(&pdf, P1, wapBox, HOURS*2, DAYS*SMALL_COLS)
	// Add time scale (mark all hours)
	pdf.SetFontSize(8)
	pdf.SetFillColor(0x00, 0x00, 0x00)
	pdf.SetStrokeColor(0x00, 0x00, 0x00)
	for hour := wap.dayStart.Hour(); hour <= wap.dayEnd.Hour(); hour += 1 {
		p := Add(ToGridSystem(DayTime(hour, 0), 0), gopdf.Point{X: -20, Y: -6})
		pdf.SetX(p.X)
		pdf.SetY(p.Y)
		// convert to military time format
		pdf.Cell(nil, fmt.Sprintf("%02d00", hour))
	}

	// TODO repeating tasks
	// TODO columns
	for _, event := range wap.events {
		// A test rectangle
		cat := event.json.Category
		if cat == nil {
			pdf.SetFillColor(127, 127, 127)
		} else if c, ok := wap.colors[*cat]; ok {
			pdf.SetFillColor(c.R, c.G, c.B)
		} else {
			pdf.SetFillColor(127, 127, 127)
		}
		RectStart := ToGridSystem(event.start, event.dayOffset)
		minutes := event.end.Sub(event.start).Minutes()
		PrintRect(&pdf, RectStart, colWidth, minutes*minuteHeight)
		// TODO Add the text
	}

	// possibly add more pages
	pdf.WritePdf(outputPath)
	return nil
}

func PrintRect(pdf *gopdf.GoPdf, p gopdf.Point, width, height float64) {
	err := pdf.Rectangle(p.X, p.Y, p.X+width, p.Y+height, "DF", 0, 0)
	if err != nil {
		log.Println("ERROR", err)
	}
}

func Add(p1, p2 gopdf.Point) gopdf.Point {
	return gopdf.Point{X: p1.X + p2.X, Y: p1.Y + p2.Y}
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
}
