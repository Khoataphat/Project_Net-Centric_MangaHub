package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mangahub/internal/models"
	"mangahub/internal/protocols/udp"
	"mangahub/proto"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	httpBaseURL = "http://127.0.0.1:8080/api"
	tcpAddress  = "127.0.0.1:9090"
	grpcAddress = "127.0.0.1:50051"
	wsChatURL   = "ws://127.0.0.1:8080/api/ws/chat"
	udpAddress  = "127.0.0.1:8888"
)

func main() {
	fmt.Println("=== MANGAHUB CLI CLIENT ===")
	fmt.Println("Hỗ trợ 5 giao thức: HTTP, TCP, GRPC, WS, UDP")
	fmt.Println(`Gõ "help" để xem ví dụ, "exit" để thoát.`)

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("mangahub> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		args, err := parseArgs(input)
		if err != nil {
			fmt.Printf("Lỗi cú pháp: %v\n", err)
			continue
		}
		if len(args) == 0 {
			continue
		}

		switch strings.ToLower(args[0]) {
		case "exit", "quit":
			fmt.Println("Đang thoát...")
			return
		case "help":
			printHelp()
		case "http":
			handleHTTP(args[1:])
		case "tcp":
			handleTCP(args[1:])
		case "grpc":
			handleGRPC(args[1:])
		case "ws":
			handleWS(args[1:])
		case "udp":
			handleUDP(args[1:])
		default:
			fmt.Println("Lệnh không hợp lệ. Hỗ trợ: http, tcp, grpc, ws, udp, help, exit")
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Lỗi đọc input: %v\n", err)
	}
}

func printHelp() {
	fmt.Println("Ví dụ lệnh:")
	fmt.Println("  http register admin 123")
	fmt.Println("  http login admin 123")
	fmt.Println(`  http mangas "one piece"`)
	fmt.Println("  http scan 101")
	fmt.Println("  tcp sync 101 5 12")
	fmt.Println("  grpc scan 101")
	fmt.Println(`  ws chat "Hello anh em"`)
	fmt.Println("  udp ping")
	fmt.Println(`  udp notify 101 12 "Chapter mới đã ra mắt"`)
}

func handleHTTP(args []string) {
	if len(args) == 0 {
		fmt.Println("Cú pháp HTTP: register|login|mangas|scan ...")
		return
	}

	switch strings.ToLower(args[0]) {
	case "register":
		if len(args) < 3 {
			fmt.Println("Ví dụ: http register <username> <password>")
			return
		}
		payload := map[string]string{
			"username": args[1],
			"password": args[2],
		}
		sendHTTPJSON(http.MethodPost, httpBaseURL+"/register", payload)
	case "login":
		if len(args) < 3 {
			fmt.Println("Ví dụ: http login <username> <password>")
			return
		}
		payload := map[string]string{
			"username": args[1],
			"password": args[2],
		}
		sendHTTPJSON(http.MethodPost, httpBaseURL+"/login", payload)
	case "mangas":
		endpoint := httpBaseURL + "/mangas"
		if len(args) > 1 {
			params := url.Values{}
			params.Set("q", strings.Join(args[1:], " "))
			endpoint += "?" + params.Encode()
		}
		sendHTTPGet(endpoint)
	case "scan":
		if len(args) < 2 {
			fmt.Println("Ví dụ: http scan <mangaID>")
			return
		}
		endpoint := httpBaseURL + "/admin/scan-manga?id=" + url.QueryEscape(args[1])
		sendHTTPGet(endpoint)
	default:
		fmt.Println("HTTP chỉ hỗ trợ: register, login, mangas, scan")
	}
}

func handleTCP(args []string) {
	if len(args) == 0 {
		fmt.Println("Cú pháp TCP: sync <userID> <mangaID> <chapter>")
		return
	}

	switch strings.ToLower(args[0]) {
	case "sync":
		if len(args) < 4 {
			fmt.Println("Ví dụ: tcp sync <userID> <mangaID> <chapter>")
			return
		}

		userID, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Println("userID phải là số nguyên")
			return
		}
		mangaID, err := strconv.Atoi(args[2])
		if err != nil {
			fmt.Println("mangaID phải là số nguyên")
			return
		}
		chapter, err := strconv.Atoi(args[3])
		if err != nil {
			fmt.Println("chapter phải là số nguyên")
			return
		}

		conn, err := net.Dial("tcp", tcpAddress)
		if err != nil {
			fmt.Printf("[TCP] Không thể kết nối %s: %v\n", tcpAddress, err)
			return
		}
		defer conn.Close()

		authPayload := models.SyncPayload{
			Type:   "AUTH",
			UserID: userID,
		}
		updatePayload := models.SyncPayload{
			Type:    "UPDATE_PROGRESS",
			UserID:  userID,
			MangaID: mangaID,
			Chapter: chapter,
			Message: "CLI sync update",
		}

		if err := writeJSONLine(conn, authPayload); err != nil {
			fmt.Printf("[TCP] Lỗi gửi AUTH: %v\n", err)
			return
		}
		if err := writeJSONLine(conn, updatePayload); err != nil {
			fmt.Printf("[TCP] Lỗi gửi UPDATE_PROGRESS: %v\n", err)
			return
		}

		fmt.Printf("[TCP] Đã gửi sync cho user=%d manga=%d chapter=%d tới %s\n", userID, mangaID, chapter, tcpAddress)
	default:
		fmt.Println("TCP chỉ hỗ trợ: sync")
	}
}

func handleGRPC(args []string) {
	if len(args) == 0 {
		fmt.Println("Cú pháp gRPC: scan <mangaID>")
		return
	}
	if strings.ToLower(args[0]) != "scan" || len(args) < 2 {
		fmt.Println("Ví dụ: grpc scan <mangaID>")
		return
	}

	id, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Println("mangaID phải là số nguyên")
		return
	}

	conn, err := grpc.NewClient(grpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("[gRPC] Không thể kết nối %s: %v\n", grpcAddress, err)
		return
	}
	defer conn.Close()

	client := proto.NewMangaServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	startedAt := time.Now()
	res, err := client.GetMangaDetail(ctx, &proto.MangaRequest{Id: int32(id)})
	if err != nil {
		fmt.Printf("[gRPC] Lỗi gọi GetMangaDetail: %v\n", err)
		return
	}

	fmt.Printf("[gRPC] Phản hồi sau %dms:\n", time.Since(startedAt).Milliseconds())
	printPrettyJSON(res)
}

func handleWS(args []string) {
	if len(args) == 0 {
		fmt.Println(`Cú pháp WS: chat "nội dung"`)
		return
	}
	if strings.ToLower(args[0]) != "chat" || len(args) < 2 {
		fmt.Println(`Ví dụ: ws chat "Hello anh em"`)
		return
	}

	conn, _, err := websocket.DefaultDialer.Dial(wsChatURL, nil)
	if err != nil {
		fmt.Printf("[WS] Không thể kết nối %s: %v\n", wsChatURL, err)
		return
	}
	defer conn.Close()

	message := strings.Join(args[1:], " ")
	if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
		fmt.Printf("[WS] Lỗi gửi tin nhắn: %v\n", err)
		return
	}

	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, reply, err := conn.ReadMessage()
	if err != nil {
		fmt.Printf("[WS] Đã gửi thành công, chưa nhận phản hồi trong thời gian chờ: %v\n", err)
		return
	}

	fmt.Printf("[WS] Server broadcast: %s\n", string(reply))
}

func handleUDP(args []string) {
	if len(args) == 0 {
		fmt.Println(`Cú pháp UDP: ping | notify <mangaID> <chapter> "title"`)
		return
	}

	var payload udp.UDPPayload

	switch strings.ToLower(args[0]) {
	case "ping":
		payload = udp.UDPPayload{
			MangaID:   "cli-ping",
			Chapter:   0,
			Title:     "CLI Ping",
			Timestamp: time.Now().Unix(),
		}
	case "notify":
		if len(args) < 4 {
			fmt.Println(`Ví dụ: udp notify <mangaID> <chapter> "title"`)
			return
		}
		chapter, err := strconv.Atoi(args[2])
		if err != nil {
			fmt.Println("chapter phải là số nguyên")
			return
		}
		payload = udp.UDPPayload{
			MangaID:   args[1],
			Chapter:   chapter,
			Title:     strings.Join(args[3:], " "),
			Timestamp: time.Now().Unix(),
		}
	default:
		fmt.Println("UDP chỉ hỗ trợ: ping, notify")
		return
	}

	conn, err := net.Dial("udp", udpAddress)
	if err != nil {
		fmt.Printf("[UDP] Không thể kết nối %s: %v\n", udpAddress, err)
		return
	}
	defer conn.Close()

	body, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("[UDP] Lỗi encode payload: %v\n", err)
		return
	}

	if _, err := conn.Write(body); err != nil {
		fmt.Printf("[UDP] Lỗi gửi datagram: %v\n", err)
		return
	}

	fmt.Printf("[UDP] Đã gửi datagram tới %s\n", udpAddress)
	printPrettyJSON(payload)
}

func sendHTTPJSON(method, endpoint string, payload any) {
	body, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("[HTTP] Lỗi encode JSON: %v\n", err)
		return
	}

	req, err := http.NewRequest(method, endpoint, bytes.NewReader(body))
	if err != nil {
		fmt.Printf("[HTTP] Không thể tạo request: %v\n", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("[HTTP] Lỗi gọi %s: %v\n", endpoint, err)
		return
	}
	defer resp.Body.Close()

	printHTTPResponse(resp)
}

func sendHTTPGet(endpoint string) {
	resp, err := http.Get(endpoint)
	if err != nil {
		fmt.Printf("[HTTP] Lỗi gọi %s: %v\n", endpoint, err)
		return
	}
	defer resp.Body.Close()

	printHTTPResponse(resp)
}

func printHTTPResponse(resp *http.Response) {
	fmt.Printf("[HTTP] Status: %s\n", resp.Status)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("[HTTP] Lỗi đọc response: %v\n", err)
		return
	}

	var pretty bytes.Buffer
	if json.Indent(&pretty, body, "", "  ") == nil {
		fmt.Println(pretty.String())
		return
	}

	fmt.Println(string(body))
}

func writeJSONLine(w io.Writer, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = w.Write(data)
	return err
}

func printPrettyJSON(v any) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Printf("Lỗi format JSON: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

func parseArgs(input string) ([]string, error) {
	var (
		args     []string
		current  strings.Builder
		inQuotes bool
	)

	for i := 0; i < len(input); i++ {
		ch := input[i]

		switch ch {
		case '"':
			inQuotes = !inQuotes
		case ' ', '\t':
			if inQuotes {
				current.WriteByte(ch)
				continue
			}
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteByte(ch)
		}
	}

	if inQuotes {
		return nil, fmt.Errorf("thiếu dấu nháy kép đóng")
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args, nil
}
