// Package demo generates realistic CMCD telemetry for testing.
package demo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

var (
	bitrateLadder = []int{400, 800, 1500, 3200, 6000}
	cdns          = []string{"akamai", "cloudflare", "fastly"}
	countries     = []weightedChoice{
		{"US", 40}, {"DE", 10}, {"GB", 10}, {"FR", 10},
		{"JP", 10}, {"BR", 10}, {"AU", 5}, {"IN", 5},
	}
	formats = []string{"h", "d"} // HLS, DASH
	streams = []string{"v", "l"} // VOD, live
)

type weightedChoice struct {
	value  string
	weight int
}

func pickWeighted(choices []weightedChoice) string {
	total := 0
	for _, c := range choices {
		total += c.weight
	}
	r := rand.Intn(total)
	for _, c := range choices {
		r -= c.weight
		if r < 0 {
			return c.value
		}
	}
	return choices[0].value
}

// Config holds demo generator settings.
type Config struct {
	CollectorURL string
	Sessions     int
	Duration     time.Duration
}

// Run starts the demo traffic generator.
func Run(cfg Config) error {
	if cfg.Sessions <= 0 {
		cfg.Sessions = 50
	}
	if cfg.Duration <= 0 {
		cfg.Duration = 5 * time.Minute
	}

	log.Printf("Demo: generating %d sessions over %s → %s", cfg.Sessions, cfg.Duration, cfg.CollectorURL)

	var wg sync.WaitGroup
	for i := 0; i < cfg.Sessions; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			runSession(cfg, id)
		}(i)
		// Stagger session starts
		time.Sleep(time.Duration(rand.Intn(2000)) * time.Millisecond)
	}
	wg.Wait()
	log.Println("Demo: all sessions complete")
	return nil
}

func runSession(cfg Config, id int) {
	sessionID := fmt.Sprintf("demo-session-%04d", id)
	contentID := fmt.Sprintf("content-%03d", rand.Intn(20))
	cdn := cdns[rand.Intn(len(cdns))]
	country := pickWeighted(countries)
	sf := formats[rand.Intn(len(formats))]
	st := streams[rand.Intn(len(streams))]
	topBitrate := bitrateLadder[len(bitrateLadder)-1]

	// Session duration: 30s to 5min
	sessionDur := time.Duration(30+rand.Intn(270)) * time.Second
	if sessionDur > cfg.Duration {
		sessionDur = cfg.Duration
	}

	currentBR := bitrateLadder[1] // start at 800
	throughput := 5000 + rand.Intn(50000)
	bufferLen := 0
	startup := true
	rebufferProb := 0.05 + rand.Float64()*0.03 // 5-8%

	chunkDuration := 4000 // ms
	end := time.Now().Add(sessionDur)

	for time.Now().Before(end) {
		// Simulate ABR: climb if throughput >> bitrate, drop if congested
		if throughput > currentBR*3 {
			currentBR = nextBitrate(currentBR, 1)
		} else if throughput < currentBR*2 {
			currentBR = nextBitrate(currentBR, -1)
		}

		// Throughput variation
		throughput = int(float64(throughput) * (0.8 + rand.Float64()*0.4))
		if throughput < 500 {
			throughput = 500
		}

		bufferLen += chunkDuration
		if bufferLen > 30000 {
			bufferLen = 30000
		}

		bs := false
		if rand.Float64() < rebufferProb {
			bs = true
			bufferLen = 0
		}

		cmcdStr := fmt.Sprintf("br=%d,bl=%d,d=%d,mtp=%d,ot=v,sf=%s,st=%s,tb=%d,sid=\"%s\",cid=\"%s\"",
			currentBR, bufferLen, chunkDuration, throughput, sf, st, topBitrate, sessionID, contentID)

		if bs {
			cmcdStr += ",bs"
		}
		if startup {
			cmcdStr += ",su"
			startup = false
		}

		body := map[string]string{
			"cmcd":         cmcdStr,
			"cdn":          cdn,
			"country_code": country,
		}
		jsonBody, _ := json.Marshal(body)

		resp, err := http.Post(cfg.CollectorURL+"/v1/cmcd", "application/json", bytes.NewReader(jsonBody))
		if err != nil {
			log.Printf("Demo session %s: send error: %v", sessionID, err)
		} else {
			resp.Body.Close()
		}

		// Wait for next chunk interval (slightly randomized)
		time.Sleep(time.Duration(chunkDuration)*time.Millisecond + time.Duration(rand.Intn(500))*time.Millisecond)
	}
}

func nextBitrate(current, direction int) int {
	for i, br := range bitrateLadder {
		if br == current {
			next := i + direction
			if next < 0 {
				return bitrateLadder[0]
			}
			if next >= len(bitrateLadder) {
				return bitrateLadder[len(bitrateLadder)-1]
			}
			return bitrateLadder[next]
		}
	}
	return current
}
