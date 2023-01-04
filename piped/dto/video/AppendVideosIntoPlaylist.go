// Package video provides the Dto related to the Piped video/steam.
package video

// AppendVideosIntoPlaylist represents the request payload needed to append videos into a playlist using the Piped Api.
type AppendVideosIntoPlaylist struct {
	PlaylistId string    `json:"playlistId"`
	VideoIds   *[]string `json:"videoIds"`
}
