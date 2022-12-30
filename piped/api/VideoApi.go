package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	pipedVideoDto "piped-playfeed/piped/dto/video"
	"strings"
)

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
	return &video, nil
}

func ExtractVideoIdFromUrl(url string) string {
	return strings.Split(url, "=")[1]
}