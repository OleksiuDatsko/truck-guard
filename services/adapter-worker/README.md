# ⚙️ TruckGuard Adapter Worker

### 1. What is it?

The **Adapter Worker** is the asynchronous processing engine of the system. Written in **Python**, it acts as a "bridge" or "middleman" that consumes raw data from the Ingestor and transforms it into structured business events.

### 2. Purpose & How it Works

It operates as a background consumer that orchestrates the data pipeline:

1.  **Consume:** Listens to the `camera:raw` Redis Stream for new ingestion events.
2.  **Analyze (OCR):** Calls the **ANPR Service** to recognize license plates from the images stored in MinIO.
3.  **Parse:** Decodes hardware-specific payloads (JSON/Binary) into a unified format.
4.  **Finalize:** Sends the enriched data (Plate + Weight + Timing) to the **Core Service** for database storage.
5.  **Reliability:** Implements a DLQ (Dead Letter Queue) in Redis for failed messages.

Project structure:
- `src/clients`: API clients for Core, ANPR, and MinIO.
- `src/logic`: Payload parsers and the main event processor.
- `src/config`: Pydantic-based configuration management.

### 3. How to Run (Standalone)

#### **Prerequisites**

*   **Python 3.12+**
*   **Redis** (with active streams)
*   **Access to Core & ANPR APIs**

#### **Configuration**

Set environment variables or use a `.env` file:

```env
REDIS_ADDR=localhost:6379
CORE_API_URL=http://localhost:8080
ANPR_API_URL=http://localhost:8000
MINIO_ENDPOINT=localhost:9000
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
