package main

import (
	"context"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func ProcessIncomingEvent(file *multipart.FileHeader, deviceID, payload string) (*IngestEvent, error) {
	objectName := fmt.Sprintf("%s/%s.jpg", time.Now().Format("2006-01-02"), uuid.New().String())

	src, _ := file.Open()
	defer src.Close()
	
	_, err := MinioClient.PutObject(ctx, BucketName, objectName, src, file.Size, minio.PutObjectOptions{
		ContentType: "image/jpeg",
	})
	if err != nil { return nil, err }

	event := IngestEvent{
		DeviceID: deviceID,
		ImageKey: objectName,
		Payload:  payload,
		At:       time.Now(),
	}

	RDB.XAdd(ctx, &redis.XAddArgs{
		Stream: "camera:raw",
		Values: map[string]interface{}{"data": event.ToJSON()},
	})

	return &event, nil
}