package video

type DeletePlaylistVideoDto struct {
	PlaylistId string `json:"playlistId"`
	Index      int    `json:"index"`
}
