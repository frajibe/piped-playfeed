package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	pipedDto "piped-playfeed/piped/dto"
	pipedVideoDto "piped-playfeed/piped/dto/video"
	"strings"
	"time"
)

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

func FetchChannelVideos(channel *pipedDto.ChannelDto, oldestDateAllowed time.Time, instanceBaseUrl string) (*[]pipedVideoDto.VideoDto, error) {
	var videos []pipedVideoDto.VideoDto
	lastVideoRecent := false
	for _, video := range channel.RelatedStreams {
		if isVideoAllowed(video, oldestDateAllowed) {
			videos = append(videos, video)
			lastVideoRecent = true
		} else {
			lastVideoRecent = false
		}
	}
	if lastVideoRecent {
		paginatedVideos, err := fetchPaginatedVideos(channel.Id, oldestDateAllowed, channel.Nextpage, instanceBaseUrl)
		if err != nil {
			return nil, err
		}
		videos = append(videos, *paginatedVideos...)
	}
	return &videos, nil
}

func fetchPaginatedVideos(channelId string, oldestDateAllowed time.Time, nextPageUrl string, instanceBaseUrl string) (*[]pipedVideoDto.VideoDto, error) {
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
	var videos []pipedVideoDto.VideoDto
	var lastVideoRecent = false
	for _, video := range nextPage.RelatedStreams {
		if isVideoAllowed(video, oldestDateAllowed) {
			videos = append(videos, video)
			lastVideoRecent = true
		} else {
			lastVideoRecent = false
		}
	}
	if lastVideoRecent {
		nextVideos, err := fetchPaginatedVideos(channelId, oldestDateAllowed, nextPage.Nextpage, instanceBaseUrl)
		if err != nil {
			return nil, err
		}
		videos = append(videos, *nextVideos...)
	}
	return &videos, nil
}

func isVideoAllowed(video pipedVideoDto.VideoDto, oldestDateAllowed time.Time) bool {
	videoDate := time.UnixMilli(video.Uploaded)
	var scheduledInFuture = videoDate.After(time.Now())
	return !videoDate.Before(oldestDateAllowed) && !videoDate.Equal(oldestDateAllowed) && !scheduledInFuture
}

func ExtractIdFromUrl(url string) string {
	return strings.Split(url, "=")[1]
}

func BuildVideoUrl(videoId string) string {
	return fmt.Sprintf("watch?v=%v", videoId)
}
