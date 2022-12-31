// Package video provides the Dto related to the Piped video/steam.
package video

// NextVideosPageDto represents the request payload needed to retrieve paginated videos/streams for a channel, using the Piped Api.
type NextVideosPageDto struct {
	Nextpage       string
	RelatedStreams []RelatedStreamDto
}
