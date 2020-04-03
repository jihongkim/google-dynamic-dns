package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type Configs struct {
	Domain string `json:"domain"`
	Google struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"google"`
	IPInfo struct {
		Key string `json:"key"`
		URL string `json:"url"`
	} `json:"ipinfo"`
	MyIP string `json:"myip"`
}

func main() {
	configs, err := loadConfigs()
	if err != nil {
		handleError(`Failed to load configs`, err)
		return
	}

	hasIPChanged, configs, err := hasIPChanged(configs)
	if err != nil {
		handleError(`Failed to check if IP has changed`, err)
		return
	}

	if !hasIPChanged {
		return
	}

	updatedDNS, err := updateDNS(configs)
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
	fmt.Println(err)
}

func buildURL(configs Configs) string {
	// Refer to the following page for more information on the API:
	// https://support.google.com/domains/answer/6147083?hl=en
	url := `https://` + configs.Google.Username + `:` + configs.Google.Password
	url += `@domains.google.com/nic/update?hostname=` + configs.Domain
	url += `&myip=` + configs.MyIP
	return url
}

func updateDNS(configs Configs) (bool, error) {
	apiURL := buildURL(configs)
	return true, nil
}

func hasIPChanged(configs Configs) (bool, Configs, error) {
	response, err := http.Get(configs.IPInfo.URL + "/ip?token=" + configs.IPInfo.Key)
	if err != nil {
		return false, configs, errors.New("Could not connect to ipinfo.io")
	}
	defer response.Body.Close()

	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return false, configs, errors.New("Weird body returned from ipinfo.io")
	}

	myip := strings.TrimSpace(string(bodyBytes))
	if myip != configs.MyIP {
		configs.MyIP = myip
		return true, configs, nil
	}

	fmt.Println("Nothing to update")
	return false, configs, nil
}

func loadConfigs() (Configs, error) {
	var configs Configs

	file, err := os.Open("configs.json")
	if err != nil {
		return configs, errors.New("Please verify that configs.json file exists")
	}

	jsonParser := json.NewDecoder(file)
	if err = jsonParser.Decode(&configs); err != nil {
		return configs, errors.New("Please verify that configs.json is a valid json file")
	}

	return configs, nil
}
