package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	pipedLoginDto "github.com/frajibe/piped-playfeed/piped/dto/login"
	"io"
	"net/http"
	"time"
)

var token string

func GetToken() string {
	return token
}

func Login(username string, password string, instanceBaseUrl string) error {
	// perform the request
	var requestDto = pipedLoginDto.LoginRequestDto{
		Username: username,
		Password: password,
	}
	payload, err := json.Marshal(requestDto)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", instanceBaseUrl+"/login", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Set("content-type", "application/json")
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("invalid response '%s'", response.Status)
	}

	// parse the response to obtain the token
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	var loginRespDto pipedLoginDto.LoginResponseDto
	err = json.Unmarshal(body, &loginRespDto)
	if err != nil {
		return err
	}
	token = loginRespDto.Token
	return nil
}
