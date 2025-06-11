package main

import (
	"fmt"
	"github.com/goccy/go-yaml"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func yamlFromBytes(dat []byte) (wap *WapJson, err error) {
	v := new(WapJson)
	reader := strings.NewReader(string(dat))
	decoder := yaml.NewDecoder(reader, yaml.DisallowUnknownField())
	if err := decoder.Decode(v); err != nil {
		return nil, err
	}
	return v, nil
}

func serveWeb() {
	http.HandleFunc("/upload", handleYAMLtoPDF)
	fmt.Println("Server started at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func handleYAMLtoPDF(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Use POST method", http.StatusMethodNotAllowed)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error reading file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	yamlBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file content: "+err.Error(), http.StatusInternalServerError)
		return
	}

	wapData, err := yamlFromBytes(yamlBytes)
	if err != nil {
		log.Fatal("ERROR reading yaml: ", err.Error())
	}
	wap := NewWAP(wapData)
	var path = "/tmp/test.pdf"
	NewPDFDrawer().Draw(wap, path)

	dat, err := os.ReadFile(path)

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=output.pdf")
	w.Write(dat)
}
