package main

import (
	"log"
	"os"
	"strings"

	"github.com/goccy/go-yaml"
)

var VERSION = "v0.1"

func readYaml(inputPath string) (wap *WapJson, err error) {
	dat, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, err
	}
	v := new(WapJson)
	reader := strings.NewReader(string(dat))
	decoder := yaml.NewDecoder(reader, yaml.DisallowUnknownField())
	if err := decoder.Decode(v); err != nil {
		return nil, err
	}
	return v, nil
}

func main() {
	// read sample data
	inputPath := "../data/det6.yaml"
	log.Println("INFO reading wap definition at ", inputPath)
	wapData, err := readYaml(inputPath)
	if err != nil {
		log.Print("ERROR reading yaml: ", err.Error())
	}
	wap := NewWAP(wapData)
	NewPDFDrawer().Draw(wap, "/dev/stdout")
}
