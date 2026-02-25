package main

import (
	"encoding/binary"
	"math"
	"net"
	"time"
)

type ReceiverConfig struct {
	Addr     string
	Duration int
}

type ReceiverStats struct {
	Packets   uint64
	Bytes     uint64
	Lost      uint64
	LatencyMs float64
	JitterMs  float64
}

func runReceiver(cfg ReceiverConfig, out chan ReceiverStats) {
	udpAddr, _ := net.ResolveUDPAddr("udp", cfg.Addr)
	conn, _ := net.ListenUDP("udp", udpAddr)
	defer conn.Close()

	// IMPORTANT: large receive buffer
	conn.SetReadBuffer(4 * 1024 * 1024)

	buf := make([]byte, 65535)

	var (
		totalPkts  uint64
		totalBytes uint64
		firstSeq   uint64
		lastSeq    uint64
		gotFirst   bool

		latSum   float64
		latSumSq float64
		latCount uint64
	)

	start := time.Now()
	end := start.Add(time.Duration(cfg.Duration) * time.Second)

	// DRAIN WINDOW (critical)
	drainUntil := end.Add(800 * time.Millisecond)

	for time.Now().Before(drainUntil) {
		conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
		n, _, err := conn.ReadFromUDP(buf)

		if err != nil {
			if time.Now().After(end) {
				break // sender done, buffer drained
			}
			continue
		}

		if n < 16 {
			continue
		}

		totalPkts++
		totalBytes += uint64(n)

		seq := binary.BigEndian.Uint64(buf[0:8])
		sendTs := int64(binary.BigEndian.Uint64(buf[8:16]))
		lat := float64(time.Now().UnixNano()-sendTs) / 1e6

		latSum += lat
		latSumSq += lat * lat
		latCount++

		if !gotFirst {
			firstSeq = seq
			lastSeq = seq
			gotFirst = true
		} else if seq > lastSeq {
			lastSeq = seq
		}
	}

	var lost uint64
	if gotFirst {
		expected := lastSeq - firstSeq + 1
		if expected > totalPkts {
			lost = expected - totalPkts
		}
	}

	mean := 0.0
	jitter := 0.0
	if latCount > 0 {
		mean = latSum / float64(latCount)
		variance := (latSumSq / float64(latCount)) - (mean * mean)
		if variance < 0 {
			variance = 0
		}
		jitter = math.Sqrt(variance)
	}

	out <- ReceiverStats{
		Packets:   totalPkts,
		Bytes:     totalBytes,
		Lost:      lost,
		LatencyMs: mean,
		JitterMs:  jitter,
	}
}
