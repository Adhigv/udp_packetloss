package main

import (
	"encoding/binary"
	"net"
	"time"
)

type SenderConfig struct {
	IP       string
	PktSize  int
	RateMB   float64
	Duration int
}

type SenderStats struct {
	Packets     uint64
	Bytes       uint64
	Throughput  float64
	PktSize     int
	TargetRate  float64
}

func runSender(cfg SenderConfig) SenderStats {
	addr := cfg.IP + ":9000"

	raddr, _ := net.ResolveUDPAddr("udp", addr)
	conn, _ := net.DialUDP("udp", nil, raddr)
	defer conn.Close()

	// IMPORTANT: large send buffer
	conn.SetWriteBuffer(4 * 1024 * 1024)

	buf := make([]byte, cfg.PktSize)

	bytesPerSec := cfg.RateMB * 1e6
	bytesPerTick := bytesPerSec / 100.0 // 10ms tick
	var carry float64

	var seq uint64
	var totalPkts uint64
	var totalBytes uint64

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	start := time.Now()
	end := start.Add(time.Duration(cfg.Duration) * time.Second)

	for time.Now().Before(end) {
		<-ticker.C

		carry += bytesPerTick
		pkts := int(carry / float64(cfg.PktSize))
		carry -= float64(pkts * cfg.PktSize)

		for i := 0; i < pkts; i++ {
			binary.BigEndian.PutUint64(buf[0:8], seq)
			binary.BigEndian.PutUint64(buf[8:16], uint64(time.Now().UnixNano()))

			if n, err := conn.Write(buf); err == nil {
				totalPkts++
				totalBytes += uint64(n)
				seq++
			}
		}
	}

	elapsed := time.Since(start).Seconds()

	return SenderStats{
		Packets:    totalPkts,
		Bytes:      totalBytes,
		Throughput: float64(totalBytes) / elapsed / 1e6,
		PktSize:    cfg.PktSize,
		TargetRate: cfg.RateMB,
	}
}
