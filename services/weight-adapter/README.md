# ⚖️ TruckGuard Weight Adapter Worker

### 1. What is it?

The **Weight Adapter Worker** is an asynchronous processing service written in **Python**. It serves as a specialized handler for raw weight sensor data, bridging the gap between high-speed ingestion and the structured storage in the Core service.

### 2. Purpose & How it Works

The worker consumes data from the `weight:raw` Redis Stream and performs the following steps:

1.  **Consume**: Listens for new messages in the `weight:raw` stream.
2.  **Configuration Retrieval**: Fetches the specific scale configuration (parsing rules, field mappings) from the **Core Service**.
3.  **Parse**: Decodes the raw payload according to the defined rules for that specific scale model.
4.  **Finalize**: Sends the structured weight event (with normalized values) to the **Core Service**.
5.  **Error Handling**: Failed messages are moved to a Dead Letter Queue (`weight:dlq`) in Redis for later inspection.

### 3. How to Run (Standalone)

#### **Prerequisites**

- **Python 3.12+**
- **Redis** (with active streams)
- **Access to Core Service API**

#### **Configuration**

Set environment variables or use a `.env` file:

```env
REDIS_ADDR=localhost:6379
CORE_API_URL=http://localhost:8080
STREAM_RAW=weight:raw
STREAM_DLQ=weight:dlq
POLL_INTERVAL=5.0
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
