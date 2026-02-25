package main

import (
	"encoding/json"
	"net/http"
	"time"
)

type TestRequest struct {
	IP       string  `json:"ip"`
	PktSize  int     `json:"pkt_size"`
	RateMB   float64 `json:"rate_mb"`
	Duration int     `json:"duration"`
}

type TestResult struct {
	Sender         SenderStats   `json:"sender"`
	Receiver       ReceiverStats `json:"receiver"`
	AbsoluteLoss   uint64        `json:"absolute_loss"`
	LossPercentage float64       `json:"loss_percentage"`
}

func main() {

	http.Handle("/", http.FileServer(http.Dir("static")))
	http.HandleFunc("/start", startTest)

	println("Open http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}


func startTest(w http.ResponseWriter, r *http.Request) {
	var req TestRequest
	json.NewDecoder(r.Body).Decode(&req)

	recvCh := make(chan ReceiverStats, 1)

	go runReceiver(
		ReceiverConfig{
			Addr:     req.IP + ":9000",
			Duration: req.Duration,
		},
		recvCh,
	)

	time.Sleep(300 * time.Millisecond)

	senderStats := runSender(SenderConfig{
		IP:       req.IP,
		PktSize:  req.PktSize,
		RateMB:   req.RateMB,
		Duration: req.Duration,
	})

	receiverStats := <-recvCh

	absLoss := uint64(0)
	lossPct := 0.0
	if senderStats.Packets > receiverStats.Packets {
		absLoss = senderStats.Packets - receiverStats.Packets
		lossPct = float64(absLoss) * 100 / float64(senderStats.Packets)
	}

	json.NewEncoder(w).Encode(TestResult{
		Sender:         senderStats,
		Receiver:       receiverStats,
		AbsoluteLoss:   absLoss,
		LossPercentage: lossPct,
	})
}
