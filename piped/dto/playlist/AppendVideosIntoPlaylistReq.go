package playlist

type AppendVideosIntoPlaylistReq struct {
	PlaylistId string    `json:"playlistId"`
	VideoIds   *[]string `json:"videoIds"`
}
