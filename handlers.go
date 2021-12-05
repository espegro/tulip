package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

// Web handlers
func hello(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	fmt.Fprintf(w, "<head>Dummy</head><html>")
	fmt.Fprintf(w, "hello, %s! - lets %s \n", p.ByName("name"), p.ByName("action"))
}

// Set up new filter
func newhandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	filtername := string(p.ByName("name"))
	filtersize, err := strconv.Atoi(p.ByName("size"))
	if err != nil {
		log.Printf("Error in new function, error in size: %v\n", p.ByName("size"))
		w.WriteHeader(422)
		return
	}

	filterhash, err := strconv.Atoi(p.ByName("hash"))
	if err != nil {
		log.Printf("Error in new function, error in hash: %v\n", p.ByName("hash"))
		w.WriteHeader(422)
		return
	}

	filterdecay, err := strconv.Atoi(p.ByName("decay"))
	if err != nil {
		log.Printf("Error in new function, error in decay: %v\n", p.ByName("decay"))
		w.WriteHeader(422)
		return
	}

	filtermax, err := strconv.Atoi(p.ByName("max"))
	if err != nil {
		log.Printf("Error in new function, error in max: %v\n", p.ByName("max"))
		w.WriteHeader(422)
		return
	}

	log.Printf("Created new filter: %v, %v, %v, %v, %v\n", filtername, filtersize, filterhash, filterdecay, filtermax)
	fmt.Printf("Created new filter: %v, %v, %v, %v, %v\n", filtername, filtersize, filterhash, filterdecay, filtermax)

	// Create new filter and add to filters map
	newfilter := NewBloom(uint(filtersize), uint(filterhash), uint(filterdecay), uint(filtermax))
	newfilter.Name = filtername
	filters[filtername] = *newfilter

	// Return status
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{\n\"status\": \"OK\",\n\"name\":\"%v\"\n}", p.ByName("name"))
	fmt.Fprintf(w, "\n\n")
	return
}

// Add value to filter
func addhandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	filtername := string(p.ByName("name"))
	filtervalue := string(p.ByName("value"))
	if val, ok := filters[filtername]; ok {
		w.WriteHeader(200)
		fmt.Fprintf(w, "OK\n")
		val.Add([]byte(filtervalue))
	} else {
		w.WriteHeader(404)
		fmt.Fprintf(w, "Filter not found!\n")
	}
	return
}

// Add value to filter if not set
func addifnotsethandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	filtername := string(p.ByName("name"))
	filtervalue := string(p.ByName("value"))
	if val, ok := filters[filtername]; ok {
		w.WriteHeader(200)
		fmt.Fprintf(w, "OK\n")
		val.AddIfNotSet([]byte(filtervalue))
	} else {
		w.WriteHeader(404)
		fmt.Fprintf(w, "Filter not found!\n")
	}
	return
}

// Test if value is in filter
func testhandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	filtername := string(p.ByName("name"))
	filtervalue := string(p.ByName("value"))
	if val, ok := filters[filtername]; ok {
		if val.Test([]byte(filtervalue)) {
			w.WriteHeader(200)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, "{\n\"status\": \"OK\",\n\"found\": true\n}\n")
		} else {
			w.WriteHeader(404)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, "{\n\"status\": \"NOT FOUND\",\n\"found\": false\n}\n")
		}
	} else {
		w.WriteHeader(422)
		fmt.Fprintf(w, "Filter not found!\n")
	}
	return
}

// Reset filter to zero
func resethandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	filtername := string(p.ByName("name"))
	if val, ok := filters[filtername]; ok {
		val.Reset()
		// Set status and type
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "{\n\"status\": \"OK\",\n\"name\":\"%v\"\n", p.ByName("name"))
		fmt.Fprintf(w, "\"values\": \"%v\"\n}\n", val.Elements)
	} else {
		w.WriteHeader(404)
		fmt.Fprintf(w, "Filter not found!\n")
	}
	return
}

// Remove filter
func destroyhandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	filtername := string(p.ByName("name"))
	if _, ok := filters[filtername]; ok {
		delete(filters, filtername)
		// Set status and type
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "{\n\"status\": \"OK\",\n\"name\":\"%v\"\n}\n", p.ByName("name"))
	} else {
		w.WriteHeader(404)
		fmt.Fprintf(w, "Filter not found!\n")
	}
	return
}

// List filters
func listhandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Set status and type
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{\n\"status\": \"OK\",\n")
	fmt.Fprintf(w, "\"filters\": [ ")
	// Itearate over keys to make array
	skip := false
	for k := range filters {
		// skip first ,
		if skip {
			fmt.Fprintf(w, " , ")
		}
		fmt.Fprintf(w, "\"%v\"", k)
		skip = true
	}
	fmt.Fprintf(w, "]\n}\n")
	return
}

// Force save of state
func savehandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	savefilterstate(config.FilterStateName)
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{\n\"status\": \"OK\"\n}\n")
}

// Force load of state
func loadhandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	loadfilterstate(config.FilterStateName)
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{\n\"status\": \"OK\"\n}\n")
}

// Print debug info
func debughandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	filtername := string(p.ByName("name"))
	b, err := json.Marshal(filters[filtername])
	if err == nil {
		fmt.Fprintf(w, string(b))
	}
	return
}

// Get info on filer
func infohandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	filtername := string(p.ByName("name"))
	if val, ok := filters[filtername]; ok {
		maplock.Lock()
		size := val.Size
		nonzero := uint(0)
		for i := uint(0); i < size; i++ {
			if val.Bitset[i] > 0 {
				nonzero = nonzero + 1
			}
		}
		maplock.Unlock()
		zero := size - nonzero
		fmt.Fprintf(w, "{\n\"name\": \"%v\",\n", filtername)
		fmt.Fprintf(w, "\"size\": \"%v\",\n", size)
		fmt.Fprintf(w, "\"rounds\": \"%v\",\n", val.Rounds)
		fmt.Fprintf(w, "\"decay\": \"%v\",\n", val.Decay)
		fmt.Fprintf(w, "\"max\": \"%v\",\n", val.Max)
		fmt.Fprintf(w, "\"elements\": \"%v\",\n", val.Elements[0])
		fmt.Fprintf(w, "\"zero\": \"%v\",\n", zero)
		fmt.Fprintf(w, "\"nonzero\": \"%v\"\n", nonzero)
		fmt.Fprintf(w, "}\n")
	} else {
		w.WriteHeader(404)
		fmt.Fprintf(w, "Filter not found!\n")
	}
	return
}

// Add multiline values to filter
func posterhandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	filtername := string(p.ByName("name"))
	w.WriteHeader(200)

	if val, ok := filters[filtername]; ok {
		w.WriteHeader(200)
		defer r.Body.Close()
		resbody, _ := ioutil.ReadAll(r.Body)
		lines := bytes.Split(resbody, []byte("\n"))
		count := 0
		for _, line := range lines {
			if len(string(line)) > 0 {
				val.Add(line)
				count = count + 1
			}
		}
		fmt.Fprintf(w, "OK, added %v values to filter: %v\n", count, filtername)
		log.Printf("OK, added %v values to filter: %v\n", count, filtername)
	} else {
		w.WriteHeader(404)
		fmt.Fprintf(w, "Filter not found!\n")
	}
	return
}
