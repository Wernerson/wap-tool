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
	pageSize := gopdf.PageSizeA4Landscape
	pdf.Start(gopdf.Config{PageSize: *pageSize})
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
	pagePadding := 5.0
	pdf.AddHeader(func() {
		pdf.SetY(pagePadding)
		pdf.SetX(pagePadding)
		pdf.Cell(nil, unit)
		tm := wap.data.Meta.Title
		tmW, err := pdf.MeasureTextWidth(tm)
		check(err)
		pdf.SetX(pageSize.W/2 - tmW/2)
		pdf.CellWithOption(nil, tm, gopdf.CellOption{Align: gopdf.Center})
		tr := version
		trW, err := pdf.MeasureTextWidth(tr)
		check(err)
		pdf.SetX(pageSize.W - trW - pagePadding)
		pdf.CellWithOption(nil, tr, gopdf.CellOption{Align: gopdf.Right})
	})
	pdf.AddFooter(func() {
		footerText := "footer"
		ftH, err := pdf.MeasureCellHeightByText(footerText)
		check(err)
		pdf.SetY(pageSize.H - pagePadding - ftH)
		pdf.Cell(nil, "footer")
	})

	pdf.AddPage()
	pdf.SetY(400)
	pdf.Text("page 1 content")
	pdf.AddPage()
	pdf.SetY(400)
	pdf.Text("page 2 content")
	pdf.WritePdf(outputPath)
	return nil
}
