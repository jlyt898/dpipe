package main

import (
	"encoding/json"
	"os"
)

type jsonItem struct {
	Name    string `json:"name"`    // must be exported, else json can't unmarshal
	Pattern string `json:"pattern"` // may not be same order as json file
	Parsers []string
}

type jsonConfig []jsonItem

func (this jsonItem) hasParser(parser string) bool {
	for _, p := range this.Parsers {
		if p == parser {
			return true
		}
	}

	return false
}

func (this jsonConfig) hasParser(parser string) bool {
	for _, p := range this.parsers() {
		if p == parser {
			return true
		}
	}

	return false
}

func (this jsonConfig) parsers() []string {
	r := make([]string, 0)
	for _, item := range this {
		for _, p := range item.Parsers {
			exists := false
			for _, parser := range r {
				if p == parser {
					exists = true
					break
				}
			}

			if !exists {
				r = append(r, p)
			}
		}
	}

	return r
}

func (this jsonConfig) parsersCount() int {
	return len(this.parsers())
}

func loadJsonConfig(filename string) (config jsonConfig) {
	file, e := os.Open(filename)
	if e != nil {
		panic(e)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if e := decoder.Decode(&config); e != nil {
		panic(e)
	}

	// in test mode, add test log
	if options.test {
		config = append(config,
			jsonItem{
				Name:    "test",
				Parsers: []string{"PaymentParser", "ErrorLogParser", "MemcacheFailParser"},
				Pattern: "fixtures/*.log"})
	}

	return
}
