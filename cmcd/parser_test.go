package cmcd

import (
	"testing"
)

func TestParseEmpty(t *testing.T) {
	d, err := Parse("")
	if err != nil {
		t.Fatal(err)
	}
	if d.PlaybackRate != 1.0 {
		t.Errorf("expected default playback rate 1.0, got %f", d.PlaybackRate)
	}
}

func TestParseFullExample(t *testing.T) {
	raw := `br=3200,bs,d=4000,mtp=48100,ot=v,sid="abc-123",sf=h,st=v,su,tb=6000,bl=21300`
	d, err := Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if d.EncodedBitrate != 3200 {
		t.Errorf("br: got %d, want 3200", d.EncodedBitrate)
	}
	if !d.BufferStarvation {
		t.Error("bs: expected true")
	}
	if d.ObjectDuration != 4000 {
		t.Errorf("d: got %d, want 4000", d.ObjectDuration)
	}
	if d.MeasuredThroughput != 48100 {
		t.Errorf("mtp: got %d, want 48100", d.MeasuredThroughput)
	}
	if d.ObjectType != ObjectTypeVideo {
		t.Errorf("ot: got %q, want %q", d.ObjectType, ObjectTypeVideo)
	}
	if d.SessionID != "abc-123" {
		t.Errorf("sid: got %q, want %q", d.SessionID, "abc-123")
	}
	if d.StreamingFormat != StreamingFormatHLS {
		t.Errorf("sf: got %q, want %q", d.StreamingFormat, StreamingFormatHLS)
	}
	if d.StreamType != StreamTypeVOD {
		t.Errorf("st: got %q, want %q", d.StreamType, StreamTypeVOD)
	}
	if !d.Startup {
		t.Error("su: expected true")
	}
	if d.TopBitrate != 6000 {
		t.Errorf("tb: got %d, want 6000", d.TopBitrate)
	}
	if d.BufferLength != 21300 {
		t.Errorf("bl: got %d, want 21300", d.BufferLength)
	}
}

func TestParseQuotedStrings(t *testing.T) {
	d, err := Parse(`cid="content/video-42",sid="session,with,commas"`)
	if err != nil {
		t.Fatal(err)
	}
	if d.ContentID != "content/video-42" {
		t.Errorf("cid: got %q", d.ContentID)
	}
	if d.SessionID != "session,with,commas" {
		t.Errorf("sid: got %q", d.SessionID)
	}
}

func TestParsePlaybackRate(t *testing.T) {
	d, err := Parse("pr=2.0,br=1500")
	if err != nil {
		t.Fatal(err)
	}
	if d.PlaybackRate != 2.0 {
		t.Errorf("pr: got %f, want 2.0", d.PlaybackRate)
	}
}

func TestParseVersion(t *testing.T) {
	d, err := Parse("v=1,br=800")
	if err != nil {
		t.Fatal(err)
	}
	if d.Version != 1 {
		t.Errorf("v: got %d, want 1", d.Version)
	}
}

func TestParseEncoded(t *testing.T) {
	encoded := "br%3D3200%2Cbs%2Cd%3D4000%2Csid%3D%22test%22"
	d, err := ParseEncoded(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if d.EncodedBitrate != 3200 {
		t.Errorf("br: got %d, want 3200", d.EncodedBitrate)
	}
	if !d.BufferStarvation {
		t.Error("bs: expected true")
	}
	if d.SessionID != "test" {
		t.Errorf("sid: got %q, want %q", d.SessionID, "test")
	}
}

func TestParseUnknownKeysIgnored(t *testing.T) {
	d, err := Parse("br=1500,x-custom=42,bl=5000")
	if err != nil {
		t.Fatal(err)
	}
	if d.EncodedBitrate != 1500 {
		t.Errorf("br: got %d", d.EncodedBitrate)
	}
	if d.BufferLength != 5000 {
		t.Errorf("bl: got %d", d.BufferLength)
	}
}

func TestParseInvalidInteger(t *testing.T) {
	_, err := Parse("br=notanumber")
	if err == nil {
		t.Error("expected error for invalid integer")
	}
}

func TestParseDeadlineAndRTP(t *testing.T) {
	d, err := Parse("dl=3000,rtp=50000")
	if err != nil {
		t.Fatal(err)
	}
	if d.Deadline != 3000 {
		t.Errorf("dl: got %d, want 3000", d.Deadline)
	}
	if d.RequestedThroughput != 50000 {
		t.Errorf("rtp: got %d, want 50000", d.RequestedThroughput)
	}
}

func TestParseBooleanOnlyKeys(t *testing.T) {
	d, err := Parse("bs,su")
	if err != nil {
		t.Fatal(err)
	}
	if !d.BufferStarvation {
		t.Error("bs: expected true")
	}
	if !d.Startup {
		t.Error("su: expected true")
	}
}
