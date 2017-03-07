package lib

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func checkPage(name string, config WebpageConfig, userAgent string, useUserAgent bool) error {
	client := &http.Client{}

	req, err := http.NewRequest("GET", config.URL, nil)
	if err != nil {
		log.Fatalln(err)
	}

	if config.UserAgent != "" {
		req.Header.Set("User-Agent", config.UserAgent)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Status: %s", resp.Status)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Could not read response body: %s", err.Error())
	}

	// if we have the mode enabled to extract strings
	if len(config.Matches) > 0 {
		bodyString := string(body)

		for _, m := range config.Matches {
			if !strings.Contains(bodyString, m) {
				return fmt.Errorf("Body did not contain required string [%s]", m)
			}
		}
	}

	return nil
}

// CheckWebpage checks the given web page and returns an error if it doesn't work.
func CheckWebpage(name string, config WebpageConfig) error {
	log.Println("Checking web page", name, "-", config.URL)

	var err error

	if len(config.UserAgents) > 0 {
		for _, agent := range config.UserAgents {
			log.Println("Using user agent", agent)
			err = checkPage(name, config, agent, true)
			if err != nil {
				err = fmt.Errorf("Failed\nUser Agent: %s\nError: %s", agent, err.Error())
				break
			}
		}
	} else if config.UserAgent != "" {
		err = checkPage(name, config, config.UserAgent, true)
	} else {
		err = checkPage(name, config, "", false)
	}

	return err
}
