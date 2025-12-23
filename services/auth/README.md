# 🛡️ TruckGuard Auth Service

### 1. What is it?

The **Auth Service** is the central security "gatekeeper" for the TruckGuard ecosystem. It is a high-performance microservice written in **Go** (using the Fiber framework) designed to handle authentication for both human users and IoT devices (cameras/adapters).

### 2. Purpose & How it Works

The service ensures that only authorized entities can access the system's internal resources. It operates on two levels:

* **User Authentication:** Validates login/password and issues **JWT tokens** for the SvelteKit frontend.
* **Machine Authentication:** Validates **X-API-Keys** for cameras and ingestion adapters.
* **Nginx Integration:** Works with the Nginx `auth_request` module. Before a request reaches the backend, Nginx makes a sub-request to this service to verify the token or key.

### 3. How to Run (Standalone)

#### **Prerequisites**

* **Go** (version 1.25 or higher)
* **PostgreSQL** (with an `auth` schema)
* **Environment Variables** (see below)

#### **Configuration**

Create a `.env` file or set the following variables:

```env
PORT=8081
DB_URL=postgres://user:pass@localhost:5432/truckguard
JWT_SECRET=your_secret_key

```

#### **Run Commands**

1. **Install dependencies:**
```bash
go mod tidy

```


2. **Start the service:**
```bash
go run main.go

```


3. **Build (optional):**
```bash
go build -o auth-service main.go
./auth-service

```