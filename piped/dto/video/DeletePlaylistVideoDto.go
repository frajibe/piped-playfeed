// Package video provides the Dto related to the Piped video/steam.
package video

// DeletePlaylistVideoDto represents the request payload needed to remove a video from a playlist using the Piped Api.
type DeletePlaylistVideoDto struct {
	PlaylistId string `json:"playlistId"`
	Index      int    `json:"index"`
}
