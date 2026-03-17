# StreamLens

**Open-source, ClickHouse-powered video delivery analytics engine.**

StreamLens ingests [CTA-5004 CMCD](https://cdn.cta.tech/cta/media/media/resources/standards/pdfs/cta-5004-final.pdf) telemetry from video players, stores it in ClickHouse, and provides real-time dashboards for monitoring video QoE at scale.

## Architecture

```
Video Player → CMCD beacon → StreamLens Collector → ClickHouse → Grafana
```

## Features

- **CMCD Parser** — Full CTA-5004 parser, importable as `github.com/CodecWhiz/streamlens/cmcd`
- **HTTP Collector** — Accepts CMCD via POST (JSON) and GET (query string)
- **Batch Buffer** — Efficient batched writes to ClickHouse
- **Auto-Migration** — Creates tables and materialized views on startup
- **Materialized Views** — Pre-aggregated rebuffer rates, session stats, bitrate distributions
- **Demo Generator** — Realistic multi-session CMCD traffic for testing
- **Grafana Dashboards** — Pre-built overview dashboard included

## Quick Start

```bash
# Start ClickHouse + StreamLens + Grafana
docker compose -f deployments/docker-compose.yml up --build -d

# Generate demo data
docker compose -f deployments/docker-compose.yml exec streamlens streamlens demo

# Open Grafana at http://localhost:3000 (admin / streamlens)
```

## Using as a Go Library

StreamLens packages are importable:

```go
import (
    "github.com/CodecWhiz/streamlens/cmcd"
    "github.com/CodecWhiz/streamlens/storage"
)

// Parse CMCD data
data, err := cmcd.Parse("br=3200,bs,d=4000,sid=\"abc\"")

// Connect to ClickHouse
client, err := storage.New(storage.Config{
    Addr:     "localhost:9000",
    Database: "streamlens",
})
```

## CLI Commands

```bash
streamlens serve     # Start the CMCD collector server
streamlens migrate   # Run ClickHouse schema migrations
streamlens demo      # Generate demo CMCD traffic
```

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `CLICKHOUSE_ADDR` | `localhost:9000` | ClickHouse native protocol address |
| `CLICKHOUSE_DATABASE` | `streamlens` | Database name |
| `CLICKHOUSE_USERNAME` | `default` | ClickHouse username |
| `CLICKHOUSE_PASSWORD` | *(empty)* | ClickHouse password |
| `COLLECTOR_PORT` | `8090` | HTTP collector port |
| `BUFFER_SIZE` | `1000` | Flush buffer size |
| `FLUSH_INTERVAL` | `5s` | Flush interval |
| `COLLECTOR_URL` | `http://localhost:8090` | Demo generator target |

## Project Structure

```
cmcd/           # CMCD parser (importable)
storage/        # ClickHouse client & migrations (importable)
collector/      # HTTP collector & batch buffer
demo/           # Demo traffic generator
cmd/streamlens/ # CLI binary
deployments/    # Docker Compose
dashboards/     # Grafana dashboards
```

## StreamLens Pro

For enterprise features including multi-tenant orchestration, K8s operator, CDN connectors, session replay, and geo enrichment, see [StreamLens Pro](https://github.com/CodecWhiz/streamlens-pro).

## License

Apache License 2.0 — see [LICENSE](LICENSE).
