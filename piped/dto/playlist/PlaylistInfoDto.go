package playlist

import pipedVideoDto "piped-playfeed/piped/dto/video"

type PlaylistInfoDto struct {
	RelatedStreams []pipedVideoDto.VideoDto
}
