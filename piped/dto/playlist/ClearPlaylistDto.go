// Package playlist provides the Dto related to the Piped playlists.
package playlist

// ClearPlaylistDto represents the request payload needed to clear the content of a playlist using the Piped Api.
type ClearPlaylistDto struct {
	PlaylistId string `json:"playlistId"`
}
