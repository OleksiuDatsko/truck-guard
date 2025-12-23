package main

import (
	"os"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
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
	
	r.Run(":8082")
}