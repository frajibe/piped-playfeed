// Package playlist provides the Dto related to the Piped playlists.
package playlist

// CreatePlaylistDto represents the request payload needed to create a playlist using the Piped Api.
type CreatePlaylistDto struct {
	Name string `json:"name"`
}
