package lib

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

// CheckWebpage checks the given web page and returns an error if it doesn't work.
func CheckWebpage(config WebpageConfig) error {
	log.Println("Checking web page", config.URL)

	resp, err := http.Get(config.URL)

	if err != nil {
		return err
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
