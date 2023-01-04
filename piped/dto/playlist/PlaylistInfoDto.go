// Package playlist provides the Dto related to the Piped playlists.
package playlist

import pipedVideoDto "github.com/frajibe/piped-playfeed/piped/dto/video"

// PlaylistInfoDto represents the model of the Piped playlist short description.
type PlaylistInfoDto struct {
	RelatedStreams []pipedVideoDto.RelatedStreamDto
}
