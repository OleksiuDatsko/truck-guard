# 🧠 TruckGuard Core Service

### 1. What is it?

The **Core Service** is the central management unit of the TruckGuard system. It acts as the primary API for managing system configurations, tracking camera states, and logging high-level events.

### 2. Purpose & How it Works

It manages the persistence layer for the entire system's configuration and historical data:

*   **Configuration Manager:** Stores and serves camera configurations, presets, and system-wide settings.
*   **Event Orchestrator:** Receives processed data from cameras (via the Adapter Worker) and saves it as formal `PlateEvents` or `WeightEvents`.
*   **Integration:** Communicates with the **Auth Service** to automatically provision API keys for new cameras.

The project follows a modular Go structure:
- `src/api`: REST handlers and validation middleware.
- `src/models`: Domain entities (Cameras, Presets, Events).
- `src/repository`: GORM-based data access layer.

### 3. How to Run (Standalone)

#### **Prerequisites**

*   **Go** (version 1.25 or higher)
*   **PostgreSQL**
*   **Environment Variables**

#### **Configuration**

```env
PORT=8080
DATABASE_URL=postgres://user:pass@localhost:5432/truckguard
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
    go build -o core-service
    ./core-service
    ```
