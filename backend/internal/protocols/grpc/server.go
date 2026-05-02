package grpc

import (
	"context"
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

	return &proto.MangaResponse{
		Id:          req.Id,
		Title:       title,
		Author:      author,
		Description: desc,
	}, nil
}

func StartGRPCServer(port string) {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("[gRPC] Không thể mở cổng lắng nghe: %v", err)
	}

	s := grpc.NewServer()

	proto.RegisterMangaServiceServer(s, &MangaServer{})

	log.Printf("[gRPC] Server đang chạy tại port %s", port)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("[gRPC] Lỗi khi chạy server: %v", err)
	}
}
