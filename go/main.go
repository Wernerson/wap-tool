package main

import (
	"log"
	"os"
	"strings"

	"github.com/goccy/go-yaml"
)

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
	wapData, err := readYaml("../data/det6week1.yaml")
	if err != nil {
		log.Print("Error reading yaml: ", err.Error())
	}
	wap := NewWAP(wapData)

	log.Printf("WAP: %v", wap)
}
