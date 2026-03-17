// Package storage provides ClickHouse connectivity and schema management.
package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/CodecWhiz/streamlens/internal/cmcd"
)

// Config holds ClickHouse connection parameters.
type Config struct {
	Addr     string
	Database string
	Username string
	Password string
}

// Client wraps a ClickHouse connection.
type Client struct {
	conn driver.Conn
	db   string
}

// New creates a new ClickHouse client.
func New(cfg Config) (*Client, error) {
	if cfg.Database == "" {
		cfg.Database = "streamlens"
	}
	if cfg.Addr == "" {
		cfg.Addr = "localhost:9000"
	}

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{cfg.Addr},
		Auth: clickhouse.Auth{
			Database: cfg.Database,
			Username: cfg.Username,
			Password: cfg.Password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		DialTimeout:     5 * time.Second,
		ConnMaxLifetime: time.Hour,
	})
	if err != nil {
		return nil, fmt.Errorf("clickhouse open: %w", err)
	}

	if err := conn.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("clickhouse ping: %w", err)
	}

	return &Client{conn: conn, db: cfg.Database}, nil
}

// Close closes the ClickHouse connection.
func (c *Client) Close() error {
	return c.conn.Close()
}

// InsertEvents inserts a batch of CMCD events.
func (c *Client) InsertEvents(ctx context.Context, events []cmcd.Event) error {
	if len(events) == 0 {
		return nil
	}

	batch, err := c.conn.PrepareBatch(ctx, `INSERT INTO cmcd_events (
		timestamp, session_id, content_id,
		encoded_bitrate, buffer_length, buffer_starvation,
		object_duration, deadline, measured_throughput,
		object_type, streaming_format, stream_type,
		startup, top_bitrate, playback_rate, requested_throughput,
		client_ip, country_code, cdn
	)`)
	if err != nil {
		return fmt.Errorf("prepare batch: %w", err)
	}

	for _, e := range events {
		ts := time.UnixMilli(e.Timestamp)
		err := batch.Append(
			ts,
			e.SessionID, e.ContentID,
			uint32(e.EncodedBitrate), uint32(e.BufferLength), e.BufferStarvation,
			uint32(e.ObjectDuration), uint32(e.Deadline), uint32(e.MeasuredThroughput),
			string(e.ObjectType), string(e.StreamingFormat), string(e.StreamType),
			e.Startup, uint32(e.TopBitrate), float32(e.PlaybackRate), uint32(e.RequestedThroughput),
			e.ClientIP, e.CountryCode, e.CDN,
		)
		if err != nil {
			return fmt.Errorf("append row: %w", err)
		}
	}

	return batch.Send()
}

// Conn returns the underlying connection for advanced use.
func (c *Client) Conn() driver.Conn {
	return c.conn
}
