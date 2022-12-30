package dto

import pipedVideoDto "github.com/frajibe/piped-playfeed/piped/dto/video"

type ChannelDto struct {
	Id             string
	Nextpage       string
	RelatedStreams []pipedVideoDto.RelatedStreamDto
}
