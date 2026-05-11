package grpc

import (
	"context"
	"fmt"
	"log"
	"net"

	"mangahub/internal/database"
	"mangahub/proto"

	"google.golang.org/grpc"
)

type MangaServer struct {
	proto.UnimplementedMangaServiceServer
}

func (s *MangaServer) GetMangaDetail(ctx context.Context, req *proto.MangaRequest) (*proto.MangaResponse, error) {
	var title, author, desc string

	err := database.DB.QueryRow("SELECT title, author, description FROM mangas WHERE id = ?", req.Id).
		Scan(&title, &author, &desc)

	if err != nil {
		log.Printf("[gRPC Error] Không tìm thấy Manga ID %d: %v", req.Id, err)
		return nil, err
	}

	res := &proto.MangaResponse{
		Id:          req.Id,
		Title:       title,
		Author:      author,
		Description: desc,
	}

	log.Printf("[gRPC Success] Found Manga ID %d: %s by %s", req.Id, title, author)
	return res, nil
}

// StartGRPCServer khởi chạy gRPC Server và hỗ trợ đóng an toàn (AC1, Defensive Rules)
func StartGRPCServer(ctx context.Context, addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("[gRPC] Không thể mở cổng lắng nghe tại %s: %v", addr, err)
	}

	s := grpc.NewServer()
	proto.RegisterMangaServiceServer(s, &MangaServer{})

	log.Printf("[gRPC] Server đang chạy tại %s", addr)

	// Goroutine lắng nghe tín hiệu dừng từ Context
	go func() {
		<-ctx.Done()
		log.Println("[gRPC] Đang dừng gRPC Server...")
		s.GracefulStop()
	}()

	if err := s.Serve(lis); err != nil {
		select {
		case <-ctx.Done():
			return nil // Thoát bình thường khi Shutdown
		default:
			return fmt.Errorf("[gRPC] Lỗi khi chạy server: %v", err)
		}
	}
	return nil
}
