package video

type AppendVideosIntoPlaylist struct {
	PlaylistId string    `json:"playlistId"`
	VideoIds   *[]string `json:"videoIds"`
}
