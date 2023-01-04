// Package video provides the Dto related to the Piped video/steam.
package video

// RelatedStreamDto represents the model of the Piped video/steam short description.
type RelatedStreamDto struct {
	Uploaded int64
	Url      string
	Views    int64
}
