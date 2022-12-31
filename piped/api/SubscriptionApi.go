// Package api provides functions to easily access the Piped Api.
package api

import (
	"encoding/json"
	"fmt"
	pipedDto "github.com/frajibe/piped-playfeed/piped/dto"
	"io"
	"net/http"
	"time"
)

// FetchSubscriptions calls the remote Piped instance to return the subscriptions associated with the user token.
//
// Error is returned if the call failed.
func FetchSubscriptions(instanceBaseUrl string, userToken string) (*[]pipedDto.SubscriptionDto, error) {
	// perform the request
	req, err := http.NewRequest("GET", instanceBaseUrl+"/subscriptions", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", userToken)
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("invalid response '%s'", response.Status)
	}

	// parse the response to obtain the subscriptions
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var subscriptions []pipedDto.SubscriptionDto
	err = json.Unmarshal(body, &subscriptions)
	if err != nil {
		return nil, err
	}
	return &subscriptions, nil
}
