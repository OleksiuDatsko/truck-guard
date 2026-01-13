package repository

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/redis/go-redis/v9"
	"github.com/truckguard/ingestor/src/models"
)

var (
	RDB         *redis.Client
	MinioClient *minio.Client
	BucketName  string = os.Getenv("BUCKET_NAME")
	ctx                = context.Background()
)

func InitRedis(addr string) {
	RDB = redis.NewClient(&redis.Options{Addr: addr})
}

func InitMinio(endpoint, accessKey, secretKey string) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize Minio: %v", err))
	}
	MinioClient = client
}

func ProcessIncomingEvent(file *multipart.FileHeader, deviceID, payload, sourceID, sourceName, eventType, stream string) (*models.IngestEvent, error) {
	var imageKey *string
	if file != nil {
		objectName := fmt.Sprintf("%s/%s.jpg", time.Now().Format("2006-01-02"), uuid.New().String())
		src, _ := file.Open()
		defer src.Close()

		_, err := MinioClient.PutObject(ctx, BucketName, objectName, src, file.Size, minio.PutObjectOptions{
			ContentType: "image/jpeg",
			UserMetadata: map[string]string{
				"x-amz-meta-source-id":   sourceID,
				"x-amz-meta-source-name": sourceName,
			},
		})
		if err != nil {
			return nil, err
		}
		imageKey = &objectName
	}

	event := models.IngestEvent{
		Type:       eventType,
		SourceID:   sourceID,
		SourceName: sourceName,
		DeviceID:   deviceID,
		ImageKey:   imageKey,
		Payload:    payload,
		At:         time.Now(),
	}

	_, err := RDB.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		Values: map[string]interface{}{"data": event.ToJSON()},
	}).Result()

	return &event, err
}
