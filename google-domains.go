package main

import (
	"encoding/json"
	"errors"
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
	Mode string `json:"mode"`
}

func (configs *Configs) loadConfigs() error {
	file, err := os.Open(configsPath)
	if err != nil {
		return errors.New(`Please verify that configs.json file exists`)
	}

	jsonParser := json.NewDecoder(file)
	if err = jsonParser.Decode(&configs); err != nil {
		return errors.New(`Please verify that configs.json is a valid json file`)
	}

	if configs.isDevMode() {
		configs.MyIP = `Test IP Address`
	}

	return nil
}

func (configs *Configs) hasIPChanged() (bool, error) {
	response, err := http.Get(configs.IPInfo.URL + `/ip?token=` + configs.IPInfo.Key)
	if err != nil {
		return false, errors.New(`Could not connect to ipinfo.io`)
	}
	defer response.Body.Close()

	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return false, errors.New(`Weird body returned from ipinfo.io`)
	}

	myip := strings.TrimSpace(string(bodyBytes))
	if myip != configs.MyIP {
		configs.MyIP = myip
		return true, nil
	}

	return false, nil
}

func (configs *Configs) buildURL() string {
	// Refer to the following page for more information on the API:
	// https://support.google.com/domains/answer/6147083?hl=en
	url := `https://` + configs.Google.Username + `:` + configs.Google.Password
	url += `@domains.google.com/nic/update?hostname=` + configs.Domain
	url += `&myip=` + configs.MyIP
	return url
}

func (configs *Configs) updateDNS() (bool, error) {
	apiURL := configs.buildURL()

	if configs.isDevMode() {
		return configs.parseResponse(`test`)
	}

	request, err := http.NewRequest(`POST`, apiURL, nil)
	if err != nil {
		return false, errors.New(`Could not connect to Google Domains`)
	}
	request.Header.Set(`User-Agent`, `Dynamic-DNS-Updater`)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return false, errors.New(`Could not connect to Google Domains`)
	}
	defer response.Body.Close()

	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return false, errors.New(`Weird body returned from Google Domains`)
	}

	responseBody := strings.TrimSpace(string(bodyBytes))
	return configs.parseResponse(responseBody)
}

func (configs *Configs) parseResponse(response string) (bool, error) {
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

func (configs *Configs) isDevMode() bool {
	if configs.Mode == `dev` {
		return true
	}

	return false
}
