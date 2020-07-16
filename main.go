package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

var configsPath string

func main() {
	configs := Configs{}

	absolutePath, err := filepath.Abs(`./configs.json`)
	if err != nil {
		handleError(`Failed to grab the absolute filepath for configs.json`, err)
	}
	configsPath = absolutePath

	err = configs.loadConfigs()
	if err != nil {
		handleError(`Failed to load configs`, err)
		return
	}

	hasIPChanged, err := configs.hasIPChanged()
	if err != nil {
		handleError(`Failed to check if IP has changed`, err)
		return
	}

	if !hasIPChanged {
		return
	}

	updatedDNS, err := configs.updateDNS()
	if err != nil {
		handleError(`Failed to update DNS`, err)
		return
	}

	if !updatedDNS {
		return
	}

	file, err := json.MarshalIndent(configs, "", "	")
	if err != nil {
		handleError(`Failed to generate file`, err)
		return
	}

	err = ioutil.WriteFile("configs.json", file, 0644)
	if err != nil {
		handleError(`Failed to write the new IP to the json file`, err)
		return
	}
}

func handleError(message string, err error) {
	errorsPath, errPath := filepath.Abs(`./errors.log`)
	if errPath != nil {
		err = errPath
	}

	file, _ := os.OpenFile(errorsPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)

	defer file.Close()

	log.SetOutput(file)
	log.Println(message, `-`, err)
}
