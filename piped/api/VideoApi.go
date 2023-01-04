// Package api provides functions to easily access the Piped Api.
package api

import (
	"encoding/json"
	"fmt"
	pipedVideoDto "github.com/frajibe/piped-playfeed/piped/dto/video"
	"io"
	"net/http"
	"strings"
)

// FetchVideo calls the remote Piped instance and returns the video corresponding to video metadata.
//
// Error is returned if the call failed.
func FetchVideo(videoMeta pipedVideoDto.RelatedStreamDto, instanceBaseUrl string) (*pipedVideoDto.StreamDto, error) {
	// perform the request
	response, err := http.Get(instanceBaseUrl + "/streams/" + ExtractVideoIdFromUrl(videoMeta.Url))
	if err != nil {
		return nil, err
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("invalid response '%s'", response.Status)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var video pipedVideoDto.StreamDto
	err = json.Unmarshal(body, &video)
	if err != nil {
		return nil, err
	}
	video.Url = videoMeta.Url
	video.Uploaded = videoMeta.Uploaded
	return &video, nil
}

// ExtractVideoIdFromUrl returns the video id corresponding to a video url.
//
// Example:
//
//	url='/watch?v=123-456789' -> id='123-456789'
func ExtractVideoIdFromUrl(url string) string {
	return strings.Split(url, "=")[1]
}
