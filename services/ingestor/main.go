package main

import (
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/redis/go-redis/v9"
	"os"
)

var (
	RDB         *redis.Client
	MinioClient *minio.Client
	BucketName  string = os.Getenv("BUCKET_NAME")
)

func main() {
	RDB = redis.NewClient(&redis.Options{Addr: os.Getenv("REDIS_ADDR")})

	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	secretKey := os.Getenv("MINIO_SECRET_KEY")
	minioClient, _ := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	MinioClient = minioClient

	r := gin.Default()
	r.POST("/ingest", HandleIngest)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.Run(":8082")
}
