package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/CodecWhiz/streamlens/internal/collector"
	"github.com/CodecWhiz/streamlens/internal/demo"
	"github.com/CodecWhiz/streamlens/internal/storage"
	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "streamlens",
		Short: "Open-source video delivery analytics",
		Long:  "StreamLens — ClickHouse-powered, CMCD-native video delivery analytics.",
	}

	root.AddCommand(serveCmd(), migrateCmd(), demoCmd())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func serveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start the CMCD collector server",
		RunE: func(cmd *cobra.Command, args []string) error {
			ch, err := newCHClient()
			if err != nil {
				return fmt.Errorf("clickhouse: %w", err)
			}
			defer ch.Close()

			if err := ch.Migrate(context.Background()); err != nil {
				return fmt.Errorf("migrate: %w", err)
			}

			buf := collector.NewBuffer(ch, envInt("BUFFER_SIZE", 1000), envDur("FLUSH_INTERVAL", 5*time.Second))
			defer buf.Close()

			srv := collector.NewServer(buf, envInt("COLLECTOR_PORT", 8090))

			// Graceful shutdown
			sig := make(chan os.Signal, 1)
			signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
			go func() {
				<-sig
				log.Println("Shutting down...")
				srv.Close()
			}()

			return srv.ListenAndServe()
		},
	}
}

func migrateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Run ClickHouse schema migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			ch, err := newCHClient()
			if err != nil {
				return err
			}
			defer ch.Close()
			return ch.Migrate(context.Background())
		},
	}
}

func demoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "demo",
		Short: "Generate demo CMCD traffic",
		RunE: func(cmd *cobra.Command, args []string) error {
			url := envStr("COLLECTOR_URL", "http://localhost:8090")
			return demo.Run(demo.Config{
				CollectorURL: url,
				Sessions:     envInt("DEMO_SESSIONS", 50),
				Duration:     envDur("DEMO_DURATION", 5*time.Minute),
			})
		},
	}
}

func newCHClient() (*storage.Client, error) {
	return storage.New(storage.Config{
		Addr:     envStr("CLICKHOUSE_ADDR", "localhost:9000"),
		Database: envStr("CLICKHOUSE_DATABASE", "streamlens"),
		Username: envStr("CLICKHOUSE_USERNAME", "default"),
		Password: envStr("CLICKHOUSE_PASSWORD", ""),
	})
}

func envStr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func envDur(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
