# 📡 StreamLens

**Open-source video delivery analytics. ClickHouse-powered, CMCD-native.**

StreamLens ingests [CMCD (Common Media Client Data)](https://cdn.cta.tech/cta/media/media/resources/standards/pdfs/cta-5004-final.pdf) telemetry from video players, stores it in ClickHouse, and visualizes quality-of-experience metrics in Grafana dashboards — all in a single binary.

[![CI](https://github.com/CodecWhiz/streamlens/actions/workflows/ci.yml/badge.svg)](https://github.com/CodecWhiz/streamlens/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

---

## ✨ Features

- **Full CTA-5004 CMCD parser** — bitrate, buffer, starvation, throughput, session tracking
- **HTTP beacon collector** — POST JSON or GET with query params
- **ClickHouse storage** — auto-migrating schema with materialized views
- **Batch write buffer** — efficient bulk inserts, configurable flush interval
- **Grafana dashboard** — rebuffer rate, bitrate trends, CDN comparison, session drill-down
- **Demo generator** — 50 simulated sessions with realistic ABR behavior
- **Single binary** — `streamlens serve`, `streamlens demo`, `streamlens migrate`
- **Docker Compose** — one command to run everything

## 🚀 Quickstart

```bash
git clone https://github.com/CodecWhiz/streamlens.git
cd streamlens
docker compose -f deployments/docker-compose.yml up --build -d
```

Then open:
- **Grafana** → [http://localhost:3000](http://localhost:3000) (admin / streamlens)
- **Collector** → [http://localhost:8090/health](http://localhost:8090/health)

Generate demo traffic:

```bash
# From another terminal (or inside the container)
docker compose -f deployments/docker-compose.yml exec streamlens streamlens demo
```

## 🏗️ Architecture

```
                     ┌──────────────────────┐
  Video Player ─────▶│                      │
   (CMCD beacon)     │    StreamLens        │
                     │    Collector         │──▶ ClickHouse ──▶ Grafana
  Demo Generator ───▶│    (single Go bin)   │
                     │                      │
                     └──────────────────────┘

  POST /v1/cmcd   { "cmcd": "br=3200,bs,d=4000,sid=\"abc\"" }
  GET  /v1/cmcd?CMCD=br%3D3200%2Cbs%2Cd%3D4000
  GET  /health
```

## 📊 What is CMCD?

[CTA-5004](https://cdn.cta.tech/cta/media/media/resources/standards/pdfs/cta-5004-final.pdf) (Common Media Client Data) is an industry standard where video players send delivery metadata with every chunk request. Players like **hls.js**, **dash.js**, **Shaka Player**, and **ExoPlayer** support it natively.

| Key | Meaning | Type |
|-----|---------|------|
| `br` | Encoded bitrate (kbps) | Integer |
| `bl` | Buffer length (ms) | Integer |
| `bs` | Buffer starvation | Boolean |
| `d` | Object duration (ms) | Integer |
| `mtp` | Measured throughput (kbps) | Integer |
| `ot` | Object type (v/a/av/i/m) | Token |
| `sf` | Streaming format (h/d) | Token |
| `st` | Stream type (v/l) | Token |
| `su` | Startup | Boolean |
| `sid` | Session ID | String |
| `cid` | Content ID | String |
| `tb` | Top bitrate (kbps) | Integer |

## ⚙️ Configuration

All configuration via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `CLICKHOUSE_ADDR` | `localhost:9000` | ClickHouse native protocol address |
| `CLICKHOUSE_DATABASE` | `streamlens` | Database name |
| `CLICKHOUSE_USERNAME` | `default` | ClickHouse username |
| `CLICKHOUSE_PASSWORD` | *(empty)* | ClickHouse password |
| `COLLECTOR_PORT` | `8090` | HTTP collector listen port |
| `BUFFER_SIZE` | `1000` | Events per batch flush |
| `FLUSH_INTERVAL` | `5s` | Maximum time between flushes |
| `COLLECTOR_URL` | `http://localhost:8090` | Target URL for demo generator |
| `DEMO_SESSIONS` | `50` | Number of simulated sessions |
| `DEMO_DURATION` | `5m` | Duration of demo generation |

## 🔧 Development

```bash
# Build
make build

# Run tests
make test

# Run locally (requires ClickHouse)
make run

# Generate demo traffic
make demo
```

## 📈 Dashboard Panels

The included Grafana dashboard provides:

- **Active Sessions** — unique sessions in last 5 minutes
- **Rebuffer Rate Over Time** — the #1 QoE metric, per minute
- **Average Bitrate Over Time** — quality trend
- **Throughput Distribution** — histogram of measured throughput
- **CDN Comparison** — rebuffer rate and bitrate by CDN provider
- **Top Rebuffering Sessions** — table of worst-performing sessions

## 🤝 Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing`)
5. Open a Pull Request

## 📜 License

Apache License 2.0 — see [LICENSE](LICENSE) for details.
