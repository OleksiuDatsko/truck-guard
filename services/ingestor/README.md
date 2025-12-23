# 🚀 TruckGuard Ingestor Service

### 1. What is it?
A high-performance gateway for IoT data ingestion in the TruckGuard ecosystem.

### 2. Purpose & How it Works
- **Fast Response:** Instantly acknowledges camera data (Status 202).
- **Blob Storage:** Saves raw images to **MinIO** (S3).
- **Async Processing:** Pushes events to **Redis Streams** for the Normalizer and ANPR services.

### 3. How to Run
#### **Prerequisites**
- Go 1.25+
- Running MinIO & Redis
- `raw-images` bucket created in MinIO

#### **Commands**
1. `go mod tidy`
2. `export REDIS_ADDR=localhost:6379 MINIO_ENDPOINT=localhost:9000 ...`
3. `go run main.go`