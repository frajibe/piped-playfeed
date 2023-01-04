// Package dto provides the Dto related to the Piped Api.
package dto

import pipedVideoDto "github.com/frajibe/piped-playfeed/piped/dto/video"

// ChannelDto represents the model of the Piped channel.
type ChannelDto struct {
	Id             string
	Nextpage       string
	RelatedStreams []pipedVideoDto.RelatedStreamDto
}
