package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

// File functions
func savefilterstate(statename string) {
	log.Printf("Saving state to file: %v\n", statename)
	maplock.Lock()
	b, err := json.Marshal(filters)
	if err == nil {
		ioutil.WriteFile(statename, b, os.ModePerm)
	}
	maplock.Unlock()
	return
}

func loadfilterstate(statename string) {
	log.Printf("Loading state from file: %v\n", statename)
	maplock.Lock()
	b, err := ioutil.ReadFile(statename)
	if err != nil {
		log.Print(err)
		return
	}
	json.Unmarshal(b, &filters)
	maplock.Unlock()
	return
}

func touchfile(name string) error {
	file, err := os.OpenFile(name, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	return file.Close()
}
