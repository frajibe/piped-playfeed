package playlist

import pipedVideoDto "github.com/frajibe/piped-playfeed/piped/dto/video"

type PlaylistInfoDto struct {
	RelatedStreams []pipedVideoDto.RelatedStreamDto
}
