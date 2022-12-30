package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	pipedPlaylistDto "piped-playfeed/piped/dto/playlist"
	pipedVideoDto "piped-playfeed/piped/dto/video"
	"piped-playfeed/utils"
	"time"
)

func FetchPlaylists(instanceBaseUrl string, userToken string) (*[]pipedPlaylistDto.PlaylistDto, error) {
	// perform the request
	req, err := http.NewRequest("GET", instanceBaseUrl+"/user/playlists/", nil)
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

	// parse the response to obtain the playlists
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var playlists []pipedPlaylistDto.PlaylistDto
	err = json.Unmarshal(body, &playlists)
	if err != nil {
		return nil, err
	}
	return &playlists, nil
}

func FetchPlaylistVideos(playlistId string, instanceBaseUrl string, userToken string) (*[]pipedVideoDto.RelatedStreamDto, error) {
	// perform the request
	req, err := http.NewRequest("GET", instanceBaseUrl+"/playlists/"+playlistId, nil)
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

	// parse the response to obtain the videos
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var playlistInfo pipedPlaylistDto.PlaylistInfoDto
	err = json.Unmarshal(body, &playlistInfo)
	if err != nil {
		return nil, err
	}
	return &playlistInfo.RelatedStreams, nil
}

func CreatePlaylist(name string, instanceBaseUrl string, userToken string) (*pipedPlaylistDto.CreatedPlaylistDto, error) {
	// perform the request
	var requestDto = pipedPlaylistDto.CreatePlaylistDto{
		Name: name,
	}
	payload, err := json.Marshal(requestDto)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", instanceBaseUrl+"/user/playlists/create", bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", userToken)
	req.Header.Set("content-type", "application/json")
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

	// parse the response to obtain the playlist
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var playlist pipedPlaylistDto.CreatedPlaylistDto
	err = json.Unmarshal(body, &playlist)
	if err != nil {
		return nil, err
	}
	return &playlist, nil
}

func AddVideosIntoPlaylist(playlistId string, videoIds *[]string, instanceBaseUrl string, userToken string) error {
	var requestDto = pipedVideoDto.AppendVideosIntoPlaylist{
		PlaylistId: playlistId,
		VideoIds:   videoIds,
	}
	payload, err := json.Marshal(requestDto)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", instanceBaseUrl+"/user/playlists/add", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", userToken)
	req.Header.Set("content-type", "application/json")
	client := &http.Client{
		Timeout: time.Second * 180,
	}
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("invalid response '%s'", response.Status)
	}
	return nil
}

func RemoveAllPlaylistVideos(playlistId string, instanceBaseUrl string, userToken string) error {
	playlistVideos, err := FetchPlaylistVideos(playlistId, instanceBaseUrl, userToken)
	if err != nil {
		return utils.WrapError("unable to retrieve the playlists videos", err)
	}
	playlistVideosCount := len(*playlistVideos)
	for i := 0; i < playlistVideosCount; i++ {
		var requestDto = pipedVideoDto.DeletePlaylistVideoDto{
			PlaylistId: playlistId,
			Index:      0,
		}
		payload, err := json.Marshal(requestDto)
		if err != nil {
			return err
		}
		req, err := http.NewRequest("POST", instanceBaseUrl+"/user/playlists/remove", bytes.NewBuffer(payload))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", userToken)
		req.Header.Set("content-type", "application/json")
		client := &http.Client{
			Timeout: time.Second * 30,
		}
		response, err := client.Do(req)
		if err != nil {
			return err
		}
		if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
			return fmt.Errorf("invalid response '%s'", response.Status)
		}
	}
	return nil
}
