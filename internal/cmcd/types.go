// Package cmcd implements a parser for CTA-5004 Common Media Client Data.
package cmcd

// ObjectType represents the CMCD object type token.
type ObjectType string

const (
	ObjectTypeVideo       ObjectType = "v"
	ObjectTypeAudio       ObjectType = "a"
	ObjectTypeMuxed       ObjectType = "av"
	ObjectTypeInit        ObjectType = "i"
	ObjectTypeCaption     ObjectType = "c"
	ObjectTypeTTML        ObjectType = "tt"
	ObjectTypeKey         ObjectType = "k"
	ObjectTypeOther       ObjectType = "o"
	ObjectTypeManifest    ObjectType = "m"
)

// StreamingFormat represents the streaming format token.
type StreamingFormat string

const (
	StreamingFormatHLS  StreamingFormat = "h"
	StreamingFormatDASH StreamingFormat = "d"
	StreamingFormatSS   StreamingFormat = "s"
	StreamingFormatOther StreamingFormat = "o"
)

// StreamType represents the stream type token.
type StreamType string

const (
	StreamTypeVOD  StreamType = "v"
	StreamTypeLive StreamType = "l"
)

// Data holds all parsed CMCD key-value pairs from a single request.
type Data struct {
	// Encoded bitrate in kbps (br)
	EncodedBitrate int
	// Buffer length in ms (bl)
	BufferLength int
	// Buffer starvation flag (bs)
	BufferStarvation bool
	// Object duration in ms (d)
	ObjectDuration int
	// Deadline in ms (dl)
	Deadline int
	// Measured throughput in kbps (mtp)
	MeasuredThroughput int
	// Object type (ot)
	ObjectType ObjectType
	// Streaming format (sf)
	StreamingFormat StreamingFormat
	// Stream type (st)
	StreamType StreamType
	// Startup flag (su)
	Startup bool
	// Top bitrate in kbps (tb)
	TopBitrate int
	// Playback rate (pr)
	PlaybackRate float64
	// Requested maximum throughput in kbps (rtp)
	RequestedThroughput int
	// Content ID (cid)
	ContentID string
	// Session ID (sid)
	SessionID string
	// CMCD version (v)
	Version int
}

// Event wraps CMCD data with server-side enrichment fields.
type Event struct {
	Data
	Timestamp   int64  // Unix milliseconds
	ClientIP    string
	CountryCode string
	CDN         string
}
