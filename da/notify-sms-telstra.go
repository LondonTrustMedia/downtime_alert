package da

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// TelstraAuthResponse is the authentication response JSON structure we get back from the API.
type TelstraAuthResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   string `json:"expires_in"`
}

// TelstraTextMessage represents a text message to be sent over the Telstra network.
type TelstraTextMessage struct {
	To   string `json:"to"`
	Body string `json:"body"`
}

// ToJSON converts a TelstraTextMessage to the JSON form needed to send it to the API.
func (ttm *TelstraTextMessage) ToJSON() (string, error) {
	messageBytes, err := json.Marshal(ttm)
	messageString := string(messageBytes)
	return messageString, err
}

// SendSMSTelstra sends a message over Telstra's SMS network to the specified number.
func SendSMSTelstra(consumerKey string, consumerSecret string, number string, message string) error {
	// get authorization token
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.telstra.com/v1/oauth/token?client_id=%s&client_secret=%s&grant_type=client_credentials&scope=SMS", consumerKey, consumerSecret), nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := http.DefaultClient.Do(req)

	if err != nil {
		return fmt.Errorf("Retrieving auth token failed: %s", err.Error())
	}

	// parse out auth token
	var authResponse TelstraAuthResponse
	body := make([]byte, 2048)
	bodyLength, err := response.Body.Read(body)
	if err != nil && err != io.EOF {
		return fmt.Errorf("Reading response into body failed: %s", err.Error())
	}

	err = json.Unmarshal(body[:bodyLength], &authResponse)

	if err != nil {
		return fmt.Errorf("Parsing auth response failed: %s", err.Error())
	}
	if authResponse.AccessToken == "" {
		return errors.New("Access Token was empty")
	}

	// try to send an SMS with the Telstra SMS API
	var assembledMessage TelstraTextMessage
	assembledMessage.To = number
	assembledMessage.Body = message
	assembledMessageText, err := assembledMessage.ToJSON()

	if err != nil {
		return fmt.Errorf("Failed to assemble message: %s", err.Error())
	}

	req, err = http.NewRequest("POST", "https://api.telstra.com/v1/sms/messages", strings.NewReader(assembledMessageText))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authResponse.AccessToken))

	response, err = http.DefaultClient.Do(req)

	if err != nil {
		return fmt.Errorf("Failed to send message: %s", err.Error())
	}

	return nil
}
