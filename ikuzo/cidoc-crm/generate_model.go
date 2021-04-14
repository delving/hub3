package main

import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"log"
	"os"
)

type Model struct {
	Classes    []Class    `xml:"Class" json:"classes,omitEmpty"`
	Properties []Property `xml:"Property" json:"properties,omitEmpty"`
}

type Class struct {
	About      string     `xml:"about,attr" json:"about,omitEmpty"`
	Labels     []Label    `xml:"label" json:"labels,omitEmpty"`
	SubClassOf []Resource `xml:"subClassOf" json:"subClassOf,omitEmpty"`
	Comment    string     `xml:`
}

type Property struct {
	About         string     `xml:"about,attr" json:"about,omitEmpty"`
	Labels        []Label    `xml:"label" json:"labels,omitEmpty"`
	Domain        []Resource `xml:"domain" json:"domain,omitEmpty"`
	Range         Resource   `xml:"range" json:"range,omitEmpty"`
	SubPropertyOf []Resource `xml:"subPropertyOf" json:"subPropertyOf,omitEmpty"`
}

type Resource struct {
	Resource string `xml:"resource,attr" json:"resource,omitEmpty"`
}

type Label struct {
	Text string `xml:",chardata" json:"text,omitEmpty"`
	Lang string `xml:"lang,attr" json:"lang,omitEmpty"`
}

func main() {
	xmlFile, err := os.Open("./model.rdf")
	// if we os.Open returns an error then handle it
	if err != nil {
		log.Fatal(err)
	}
	byteValue, err := ioutil.ReadAll(xmlFile)
	if err != nil {
		xmlFile.Close()
		log.Fatal(err)
	}

	// defer the closing of our xmlFile so that we can parse it later on
	model := &Model{}
	err = xml.Unmarshal(byteValue, model)
	if err != nil {
		xmlFile.Close()
		log.Fatal(err)
	}
	xmlFile.Close()

	json, _ := json.MarshalIndent(&model, "", "  ")
	println(len(json))
	err = ioutil.WriteFile("./model.json", json, 0644)
	if err != nil {
		log.Fatal(err)
	}
}
