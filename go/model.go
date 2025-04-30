package main

import (
	"fmt"
	"log"
)

type Wap struct {
	data   *WapJson
	colors map[string]RGBColor
}

func NewWAP(data *WapJson) (w *Wap) {
	w = new(Wap)
	w.data = data
	w.colors = make(map[string]RGBColor)
	w.parseColors()
	return
}

func (w Wap) String() string {
	return fmt.Sprintf("raw: %v\ncolors: %v", w.data, w.colors)
}

func (w Wap) parseColors() {
	for _, cat := range w.data.Categories {
		c, err := parseColor(*cat.Color)
		if err != nil {
			log.Println(err.Error())
			log.Println("WARNING falling back to default colors")
			// MAYBE: pick from a set of predefined columns
			c = RGBColor{127, 127, 127}
		}
		w.colors[cat.Identifier] = c
	}
}
