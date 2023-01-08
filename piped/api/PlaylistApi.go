// Package api provides functions to easily access the Piped Api.
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	pipedPlaylistDto "github.com/frajibe/piped-playfeed/piped/dto/playlist"
	pipedVideoDto "github.com/frajibe/piped-playfeed/piped/dto/video"
	"io"
	"net/http"
	"time"
)

// FetchPlaylists calls the remote Piped instance to return the playlists associated with the user token.
//
// Error is returned if the call failed.
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

// FetchPlaylistVideos calls the remote Piped instance to return the videos associated with a specific playlist.
//
// Error is returned if the call failed.
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

// CreatePlaylist calls the remote Piped instance to create a new playlist.
//
// The created playlist is returned if the call succeeded.
//
// Error is returned if the call failed.
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

// AddVideosIntoPlaylist calls the remote Piped instance to insert videos into a specific playlist.
//
// Error is returned if the call failed.
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

// ClearPlaylistVideos calls the remote Piped instance to clear a specific playlist.
//
// Error is returned if the call failed.
func ClearPlaylistVideos(playlistId string, instanceBaseUrl string, userToken string) error {
	var requestDto = pipedPlaylistDto.ClearPlaylistDto{
		PlaylistId: playlistId,
	}
	payload, err := json.Marshal(requestDto)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", instanceBaseUrl+"/user/playlists/clear", bytes.NewBuffer(payload))
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
