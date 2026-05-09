package udp

import (
	"encoding/json"
	"net"
	"testing"
	"time"
)

func TestUDPInitAndBroadcast(t *testing.T) {
	port := 12345
	
	// Test AC1 & AC4: Init Server
	err := InitUDPServer(port)
	if err != nil {
		t.Fatalf("Failed to init UDP server: %v", err)
	}

	if !GetInitializedStatus() {
		t.Error("Server should be marked as initialized")
	}

	// Test AC4: Port in use
	err = InitUDPServer(port)
	if err == nil {
		t.Error("Should return error when port is already in use")
	}

	// Setup a mock Bridge (Receiver)
	bridgeAddr := "127.0.0.1:12346"
	receiverAddr, _ := net.ResolveUDPAddr("udp", bridgeAddr)
	receiver, err := net.ListenUDP("udp", receiverAddr)
	if err != nil {
		t.Fatalf("Failed to setup mock bridge: %v", err)
	}
	defer receiver.Close()

	// Register Bridge
	AddBridge(bridgeAddr)

	// Test AC2: Payload serialization and broadcast
	testData := UDPPayload{
		MangaID:   "manga-123",
		Chapter:   5,
		Title:     "Chapter 5: The Beginning",
		Timestamp: time.Now().Unix(),
	}

	// AC3: BroadcastUpdate should not block
	start := time.Now()
	BroadcastUpdate(testData)
	duration := time.Since(start)
	
	if duration > 10*time.Millisecond {
		t.Errorf("BroadcastUpdate took too long (%v), might be blocking", duration)
	}

	// Verify data received at bridge
	buffer := make([]byte, 1024)
	receiver.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, _, err := receiver.ReadFromUDP(buffer)
	if err != nil {
		t.Fatalf("Failed to receive data at bridge: %v", err)
	}

	var receivedData UDPPayload
	err = json.Unmarshal(buffer[:n], &receivedData)
	if err != nil {
		t.Fatalf("Failed to unmarshal received data: %v", err)
	}

	if receivedData.MangaID != testData.MangaID || receivedData.Chapter != testData.Chapter {
		t.Errorf("Received data mismatch. Got %+v, want %+v", receivedData, testData)
	}
}
