package dto

import pipedVideoDto "piped-playfeed/piped/dto/video"

type ChannelDto struct {
	Id             string
	Nextpage       string
	RelatedStreams []pipedVideoDto.RelatedStreamDto
}
