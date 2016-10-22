package da

import (
	"fmt"
	"log"
	"net/http"
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

	return nil
}
