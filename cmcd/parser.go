package cmcd

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// Parse parses a CMCD query string (e.g. "br=3200,bs,d=4000,sid=\"abc\"").
// The input should be the raw CMCD value, not URL-encoded.
func Parse(raw string) (Data, error) {
	var d Data
	d.PlaybackRate = 1.0 // default per spec

	if raw == "" {
		return d, nil
	}

	pairs := splitPairs(raw)
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		key, value, hasValue := splitKV(pair)
		if err := applyField(&d, key, value, hasValue); err != nil {
			return d, fmt.Errorf("cmcd: key %q: %w", key, err)
		}
	}

	return d, nil
}

// ParseEncoded parses a URL-encoded CMCD string, as found in query parameters.
func ParseEncoded(encoded string) (Data, error) {
	decoded, err := url.QueryUnescape(encoded)
	if err != nil {
		return Data{}, fmt.Errorf("cmcd: url decode: %w", err)
	}
	return Parse(decoded)
}

// splitPairs splits the CMCD string by commas, respecting quoted strings.
func splitPairs(s string) []string {
	var pairs []string
	var current strings.Builder
	inQuote := false

	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch == '"' {
			inQuote = !inQuote
			current.WriteByte(ch)
		} else if ch == ',' && !inQuote {
			pairs = append(pairs, current.String())
			current.Reset()
		} else {
			current.WriteByte(ch)
		}
	}
	if current.Len() > 0 {
		pairs = append(pairs, current.String())
	}
	return pairs
}

// splitKV splits "key=value" or returns key with hasValue=false for boolean keys.
func splitKV(pair string) (key, value string, hasValue bool) {
	idx := strings.IndexByte(pair, '=')
	if idx < 0 {
		return pair, "", false
	}
	return pair[:idx], pair[idx+1:], true
}

func applyField(d *Data, key, value string, hasValue bool) error {
	switch key {
	case "br":
		return parseInt(value, &d.EncodedBitrate)
	case "bl":
		return parseInt(value, &d.BufferLength)
	case "bs":
		d.BufferStarvation = true
		return nil
	case "d":
		return parseInt(value, &d.ObjectDuration)
	case "dl":
		return parseInt(value, &d.Deadline)
	case "mtp":
		return parseInt(value, &d.MeasuredThroughput)
	case "ot":
		d.ObjectType = ObjectType(value)
		return nil
	case "sf":
		d.StreamingFormat = StreamingFormat(value)
		return nil
	case "st":
		d.StreamType = StreamType(value)
		return nil
	case "su":
		d.Startup = true
		return nil
	case "tb":
		return parseInt(value, &d.TopBitrate)
	case "pr":
		return parseFloat(value, &d.PlaybackRate)
	case "rtp":
		return parseInt(value, &d.RequestedThroughput)
	case "cid":
		d.ContentID = unquote(value)
		return nil
	case "sid":
		d.SessionID = unquote(value)
		return nil
	case "v":
		return parseInt(value, &d.Version)
	default:
		// Unknown keys are ignored per spec.
		return nil
	}
}

func parseInt(s string, target *int) error {
	v, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("invalid integer %q: %w", s, err)
	}
	*target = v
	return nil
}

func parseFloat(s string, target *float64) error {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return fmt.Errorf("invalid float %q: %w", s, err)
	}
	*target = v
	return nil
}

func unquote(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}
