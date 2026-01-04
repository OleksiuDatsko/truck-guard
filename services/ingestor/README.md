# 🚀 TruckGuard Ingestor Service

### 1. What is it?

The **Ingestor Service** is the high-performance entry point for IoT data in the TruckGuard ecosystem. It acts as a specialized gateway designed to handle massive streams of data from cameras and sensors with minimal latency.

### 2. Purpose & How it Works

The service is built for "fire-and-forget" high-speed ingestion:

*   **Fast Response:** Immediately acknowledges incoming multipart data (Status 202) to free up camera resources.
*   **Blob Storage:** Automatically uploads raw image frames (JPG) to **MinIO** (S3-compatible storage) for permanent archival.
*   **Async Streamer:** Pushes a lightweight event descriptor into **Redis Streams** (`camera:raw`). This triggers downstream processing by AI models and normalization workers.
*   **Authentication:** Requires a valid API Key (provided by the Auth service) for every ingestion request.

Modular structure under `src/`:
- `src/api`: Data ingestion handlers and permission middleware.
- `src/models`: Ingest event schemas.
- `src/repository`: MinIO and Redis stream drivers.

### 3. How to Run (Standalone)

#### **Prerequisites**

*   **Go** (version 1.25 or higher)
*   **Redis**
*   **MinIO** (with a bucket named via `BUCKET_NAME`)
*   **Environment Variables**

#### **Configuration**

```env
PORT=8082
REDIS_ADDR=localhost:6379
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
BUCKET_NAME=raw-images
```

#### **Run Commands**

1.  **Install dependencies:**
    ```bash
    go mod tidy
    ```

2.  **Start the service:**
    ```bash
    go run .
    ```

3.  **Build:**
    ```bash
    go build -o ingestor-service
    ./ingestor-service
    ```