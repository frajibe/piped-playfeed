// Package api provides functions to easily access the Piped Api.
package api

import (
	"encoding/json"
	"fmt"
	pipedDto "github.com/frajibe/piped-playfeed/piped/dto"
	pipedVideoDto "github.com/frajibe/piped-playfeed/piped/dto/video"
	"github.com/frajibe/piped-playfeed/utils"
	"io"
	"net/http"
	"net/url"
	"time"
)

// FetchChannel calls the remote Piped instance to return a specific channel.
//
// Error is returned if the call failed.
func FetchChannel(subscription pipedDto.SubscriptionDto, instanceBaseUrl string) (*pipedDto.ChannelDto, error) {
	response, err := http.Get(instanceBaseUrl + subscription.Url)
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
	var channel pipedDto.ChannelDto
	err = json.Unmarshal(body, &channel)
	if err != nil {
		return nil, err
	}
	return &channel, nil
}

// FetchChannelVideos calls the remote Piped instance to return the videos associated with a specific channel.
//
// Error is returned if the call failed.
func FetchChannelVideos(channel *pipedDto.ChannelDto, startDate time.Time, instanceBaseUrl string) (*[]pipedVideoDto.StreamDto, error) {
	var videos []pipedVideoDto.StreamDto
	requestNextPage := false
	for _, relatedStream := range channel.RelatedStreams {
		if relatedStream.Views >= 0 { // '= -1' if the video is scheduled in the future
			video, err := FetchVideo(relatedStream, instanceBaseUrl)
			if err != nil {
				msg := fmt.Sprintf("unable to retrieve details for the video '%s'", relatedStream.Url)
				utils.GetLoggingService().WarnFromError(utils.WrapError(msg, err))
			} else {
				if isVideoAllowed(video, startDate) {
					videos = append(videos, *video)
					requestNextPage = true
				} else {
					requestNextPage = false
					break
				}
			}
		}
	}
	if requestNextPage {
		paginatedVideos, err := fetchPaginatedVideos(channel.Id, startDate, channel.Nextpage, instanceBaseUrl)
		if err != nil {
			return nil, err
		}
		videos = append(videos, *paginatedVideos...)
	}
	return &videos, nil
}

func fetchPaginatedVideos(channelId string, startDate time.Time, nextPageUrl string, instanceBaseUrl string) (*[]pipedVideoDto.StreamDto, error) {
	// perform the request
	response, err := http.Get(instanceBaseUrl + "/nextpage/channel/" + channelId + "?nextpage=" + url.QueryEscape(nextPageUrl))
	if err != nil {
		return nil, err
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("invalid response '%s'", response.Status)
	}

	// parse the response to obtain the paginated videos
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var nextPage pipedVideoDto.NextVideosPageDto
	err = json.Unmarshal(body, &nextPage)
	if err != nil {
		return nil, err
	}
	var videos []pipedVideoDto.StreamDto
	var requestNextPage = false
	for _, relatedStream := range nextPage.RelatedStreams {
		if relatedStream.Views >= 0 { // '= -1' if the video is scheduled in the future
			video, err := FetchVideo(relatedStream, instanceBaseUrl)
			if err != nil {
				msg := fmt.Sprintf("unable to retrieve details for the video '%s'", relatedStream.Url)
				utils.GetLoggingService().WarnFromError(utils.WrapError(msg, err))
			} else {
				if isVideoAllowed(video, startDate) {
					videos = append(videos, *video)
					requestNextPage = true
				} else {
					requestNextPage = false
					break
				}
			}
		}
	}
	if requestNextPage {
		nextVideos, err := fetchPaginatedVideos(channelId, startDate, nextPage.Nextpage, instanceBaseUrl)
		if err != nil {
			return nil, err
		}
		videos = append(videos, *nextVideos...)
	}
	return &videos, nil
}

func isVideoAllowed(video *pipedVideoDto.StreamDto, startDate time.Time) bool {
	videoDate, _ := time.Parse("2006-01-02", video.UploadDate)
	return !videoDate.Before(startDate) && !videoDate.After(time.Now())
}
