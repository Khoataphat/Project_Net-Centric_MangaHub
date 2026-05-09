package udp

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
)

var (
	conn        *net.UDPConn
	bridges     = make([]*net.UDPAddr, 0)
	bridgesMu   sync.RWMutex
	initialized bool
)

// InitUDPServer khởi tạo UDP Socket lắng nghe (AC1, AC4)
func InitUDPServer(port int) error {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("không thể resolve UDP address: %v", err)
	}

	c, err := net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("lỗi khởi tạo UDP Server (port %d có thể bị chiếm dụng): %v", port, err)
	}

	conn = c
	initialized = true
	log.Printf("[UDP] Notifier Server đang lắng nghe tại port %d", port)
	
	return nil
}

// AddBridge đăng ký một node trung gian (Bridge) để nhận broadcast
func AddBridge(address string) error {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return fmt.Errorf("địa chỉ Bridge không hợp lệ: %v", err)
	}
	bridgesMu.Lock()
	defer bridgesMu.Unlock()
	bridges = append(bridges, addr)
	return nil
}

// BroadcastUpdate phát đi thông báo về chương truyện mới (AC2, AC3)
func BroadcastUpdate(data UDPPayload) {
	if !initialized {
		log.Println("[UDP] Lỗi: Cố gắng Broadcast khi Server chưa được khởi tạo")
		return
	}

	// [AC3] Chạy trong goroutine để không gây block luồng chính (REST API)
	go func() {
		// [AC2] Serialize sang JSON
		payload, err := json.Marshal(data)
		if err != nil {
			log.Printf("[UDP] Lỗi serialize payload: %v", err)
			return
		}

		bridgesMu.RLock()
		defer bridgesMu.RUnlock()

		for _, bridgeAddr := range bridges {
			_, err := conn.WriteToUDP(payload, bridgeAddr)
			if err != nil {
				log.Printf("[UDP] Lỗi gửi tới Bridge %s: %v", bridgeAddr, err)
			}
		}
	}()
}

// GetInitializedStatus trả về trạng thái khởi tạo (cho unit test)
func GetInitializedStatus() bool {
	return initialized
}
