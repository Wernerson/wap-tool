package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/goccy/go-yaml"
)

const VERSION = "v0.0.1"

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
	if len(os.Args) < 2 {
		printHelpAndExit()
	}

	var inputPath string
	var outputPath = "/dev/stdout"

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-h", "--help":
			printHelpAndExit()
		case "-o", "--output":
			if i+1 < len(args) {
				outputPath = args[i+1]
				i++
			} else {
				log.Println("ERROR: Missing value for output path")
				printHelpAndExit()
			}
		case "web":
			serveWeb()
		default:
			if inputPath == "" {
				inputPath = args[i]
			} else {
				log.Println("ERROR: Too many arguments")
				printHelpAndExit()
			}
		}
	}

	if inputPath == "" {
		log.Println("ERROR: Missing input path")
		printHelpAndExit()
	}

	log.Println("INFO reading wap definition at ", inputPath)
	wapData, err := readYaml(inputPath)
	if err != nil {
		log.Fatal("ERROR reading yaml: ", err.Error())
	}
	wap := NewWAP(wapData)
	NewPDFDrawer().Draw(wap, outputPath)
}

func printHelpAndExit() {
	fmt.Println("Usage: main <inputPath> [-o <outputPath>] [-h|--help]")
	fmt.Println("  <inputPath>       Path to the input YAML file")
	fmt.Println("  -o, --output      Output path (optional, default: /dev/stdout)")
	fmt.Println("  -h, --help        Show this help message")
	os.Exit(1)
}
