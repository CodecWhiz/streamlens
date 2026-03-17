package storage

import (
	"context"
	"fmt"
	"log"
)

// Migrate creates the database, tables, and materialized views.
func (c *Client) Migrate(ctx context.Context) error {
	log.Printf("Running migrations for database %q", c.db)

	stmts := []struct {
		name string
		sql  string
	}{
		{
			name: "create database",
			sql:  fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", c.db),
		},
		{
			name: "create cmcd_events",
			sql: fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.cmcd_events (
				timestamp DateTime64(3),
				session_id String,
				content_id String,
				encoded_bitrate UInt32,
				buffer_length UInt32,
				buffer_starvation Bool DEFAULT false,
				object_duration UInt32,
				deadline UInt32,
				measured_throughput UInt32,
				object_type LowCardinality(String),
				streaming_format LowCardinality(String),
				stream_type LowCardinality(String),
				startup Bool DEFAULT false,
				top_bitrate UInt32,
				playback_rate Float32 DEFAULT 1.0,
				requested_throughput UInt32,
				client_ip String DEFAULT '',
				country_code LowCardinality(String) DEFAULT '',
				cdn LowCardinality(String) DEFAULT '',
				ingested_at DateTime64(3) DEFAULT now64(3)
			) ENGINE = MergeTree()
			PARTITION BY toYYYYMMDD(timestamp)
			ORDER BY (session_id, timestamp)
			TTL timestamp + INTERVAL 90 DAY
			SETTINGS index_granularity = 8192`, c.db),
		},
		{
			name: "create mv_rebuffer_per_minute",
			sql: fmt.Sprintf(`CREATE MATERIALIZED VIEW IF NOT EXISTS %s.mv_rebuffer_per_minute
				ENGINE = AggregatingMergeTree()
				PARTITION BY toYYYYMMDD(minute)
				ORDER BY (minute, country_code, cdn)
				AS SELECT
					toStartOfMinute(timestamp) AS minute,
					country_code,
					cdn,
					countState() AS total,
					countIfState(buffer_starvation = true) AS rebuffers
				FROM %s.cmcd_events
				GROUP BY minute, country_code, cdn`, c.db, c.db),
		},
		{
			name: "create mv_sessions",
			sql: fmt.Sprintf(`CREATE MATERIALIZED VIEW IF NOT EXISTS %s.mv_sessions
				ENGINE = AggregatingMergeTree()
				PARTITION BY toYYYYMMDD(first_ts)
				ORDER BY session_id
				AS SELECT
					session_id,
					minState(timestamp) AS first_ts,
					maxState(timestamp) AS last_ts,
					avgState(encoded_bitrate) AS avg_bitrate,
					maxState(encoded_bitrate) AS max_bitrate,
					countIfState(buffer_starvation = true) AS rebuffer_count,
					avgState(measured_throughput) AS avg_throughput,
					countState() AS chunk_count
				FROM %s.cmcd_events
				GROUP BY session_id`, c.db, c.db),
		},
		{
			name: "create mv_bitrate_distribution",
			sql: fmt.Sprintf(`CREATE MATERIALIZED VIEW IF NOT EXISTS %s.mv_bitrate_distribution
				ENGINE = AggregatingMergeTree()
				PARTITION BY toYYYYMMDD(hour)
				ORDER BY (hour, streaming_format)
				AS SELECT
					toStartOfHour(timestamp) AS hour,
					streaming_format,
					quantileState(0.5)(encoded_bitrate) AS p50_bitrate,
					quantileState(0.95)(encoded_bitrate) AS p95_bitrate,
					avgState(encoded_bitrate) AS avg_bitrate
				FROM %s.cmcd_events
				WHERE object_type = 'v'
				GROUP BY hour, streaming_format`, c.db, c.db),
		},
	}

	for _, s := range stmts {
		log.Printf("  → %s", s.name)
		if err := c.conn.Exec(ctx, s.sql); err != nil {
			return fmt.Errorf("migration %q: %w", s.name, err)
		}
	}

	log.Println("Migrations complete")
	return nil
}
