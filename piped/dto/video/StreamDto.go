// Package video provides the Dto related to the Piped video/steam.
package video

// StreamDto represents the model of the Piped video/steam.
type StreamDto struct {
	Uploaded   int64
	UploadDate string
	Url        string
}
