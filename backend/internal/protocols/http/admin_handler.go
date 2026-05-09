package http

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"mangahub/proto"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func ScanMangaHandler(c *gin.Context) {
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID truyện không hợp lệ (phải là số)"})
		return
	}

	conn, err := grpc.NewClient("127.0.0.1:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể kết nối hệ thống gRPC"})
		return
	}
	defer conn.Close()

	client := proto.NewMangaServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	res, err := client.GetMangaDetail(ctx, &proto.MangaRequest{Id: int32(id)})

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Không tìm thấy dữ liệu gRPC",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"protocol_used": "gRPC (Binary)",
		"time_taken_ms": time.Since(time.Now()).Milliseconds(),
		"data":          res,
	})
}
