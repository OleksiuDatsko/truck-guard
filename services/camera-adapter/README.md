# ⚙️ TruckGuard Camera Adapter Worker

### 1. What is it?

The **Camera Adapter Worker** is a high-performance asynchronous processing engine written in **Python**. It acts as the primary "bridge" between raw camera ingestion and structured event logging, orchestrating AI-driven recognition and data transformation.

### 2. Purpose & How it Works

It operates as a background consumer that orchestrates the data pipeline:

1.  **Consume**: Listens to the `camera:raw` Redis Stream for new ingestion events.
2.  **Analyze (AI/OCR)**: Forwards images stored in **MinIO** to the **ANPR Service** for hardware-agnostic license plate recognition.
3.  **Parse**: Decodes manufacturer-specific payloads (JSON/XML) into the unified TruckGuard internal format.
4.  **Finalize**: Sends the enriched data (normalized Plate + Metadata) to the **Core Service** for permanent storage.
5.  **Reliability**: Implements a Dead Letter Queue (`camera:dlq`) in Redis for automatic handling of processing failures.

### 3. How to Run (Standalone)

#### **Prerequisites**

- **Python 3.12+**
- **Redis** (with active streams)
- **Access to Core & ANPR APIs**
- **Access to MinIO Storage**

#### **Configuration**

Set environment variables or use a `.env` file:

```env
REDIS_ADDR=localhost:6379
CORE_API_URL=http://localhost:8080
ANPR_API_URL=http://localhost:8000
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
...
```

#### **Run Commands**

1.  **Create virtual environment:**

    ```bash
    python -m venv venv
    source venv/bin/activate
    ```

2.  **Install dependencies:**

    ```bash
    pip install -r requirements.txt
    ```

3.  **Start the worker:**
    ```bash
    python main.py
    ```
