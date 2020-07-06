package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
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
	file, _ := os.OpenFile("errors.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)

	defer file.Close()

	log.SetOutput(file)
	log.Println(message, `-`, err)
}

func loadConfigs() (Configs, error) {
	var configs Configs

	file, err := os.Open("configs.json")
	if err != nil {
		return configs, errors.New(`Please verify that configs.json file exists`)
	}

	jsonParser := json.NewDecoder(file)
	if err = jsonParser.Decode(&configs); err != nil {
		return configs, errors.New(`Please verify that configs.json is a valid json file`)
	}

	return configs, nil
}

func hasIPChanged(configs Configs) (bool, Configs, error) {
	response, err := http.Get(configs.IPInfo.URL + "/ip?token=" + configs.IPInfo.Key)
	if err != nil {
		return false, configs, errors.New(`Could not connect to ipinfo.io`)
	}
	defer response.Body.Close()

	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return false, configs, errors.New(`Weird body returned from ipinfo.io`)
	}

	myip := strings.TrimSpace(string(bodyBytes))
	if myip != configs.MyIP {
		configs.MyIP = myip
		return true, configs, nil
	}

	return false, configs, nil
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

	request, err := http.NewRequest(`POST`, apiURL, nil)
	if err != nil {
		return false, errors.New(`Could not connect to Google Domains`)
	}
	request.Header.Set(`User-Agent`, `Dynamic-DNS-Updater`)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return false, errors.New("Could not connect to Google Domains")
	}
	defer response.Body.Close()

	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return false, errors.New("Weird body returned from Google Domains")
	}

	responseBody := strings.TrimSpace(string(bodyBytes))
	return parseResponse(responseBody)
}

func parseResponse(response string) (bool, error) {
	if strings.Contains(response, `nochg`) {
		return false, nil
	}

	if strings.Contains(response, `nohost`) {
		return false, errors.New(`The hostname does not exist, or does not have Dynamic DNS enabled`)
	}

	if strings.Contains(response, `badauth`) {
		return false, errors.New(`The username / password combination is not valid for the specified host`)
	}

	if strings.Contains(response, `notfqdn`) {
		return false, errors.New(`The supplied hostname is not a valid fully-qualified domain name`)
	}

	if strings.Contains(response, `badagent`) {
		return false, errors.New(`Your Dynamic DNS client is making bad requests. Ensure the user agent is set in the request`)
	}

	if strings.Contains(response, `abuse`) {
		return false, errors.New(`Dynamic DNS access for the hostname has been blocked due to failure to interpret previous responses correctly`)
	}

	if strings.Contains(response, `911`) {
		return false, errors.New(`An error happened on our end. Wait 5 minutes and retry`)
	}

	if strings.Contains(response, `conflict A`) {
		return false, errors.New(`A custom A or AAAA resource record conflicts with the update. Delete the indicated resource record within DNS settings page and try the update again`)
	}

	return true, nil
}
