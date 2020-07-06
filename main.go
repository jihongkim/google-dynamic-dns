package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	configs := Configs{}

	err := configs.loadConfigs()
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
	file, _ := os.OpenFile("errors.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)

	defer file.Close()

	log.SetOutput(file)
	log.Println(message, `-`, err)
}
