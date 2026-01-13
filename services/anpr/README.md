# 🤖 TruckGuard ANPR Service

### 1. What is it?

The **ANPR Service** is the "AI brain" of the TruckGuard system. It is a high-performance computer vision microservice built with **FastAPI** and **Nomeroff-Net**. It uses deep learning models (PyTorch) to detect and recognize vehicle license plates from images in real-time.

### 2. Purpose & How it Works

This service is responsible for converting raw image data into actionable text data.

* **Input:** Receives an image file via a `POST /recognize` multipart request.
* **Processing:** * Decodes the image using **OpenCV**.
* Runs the **Nomeroff-Net pipeline**:
1. **Detection:** Locates the license plate in the frame.
2. **OCR:** Reads the text from the detected plate.



* **Output:** Returns a JSON object containing the status (`found: true/false`) and a list of identified plates.
* **Security Patch:** Includes a custom `torch.load` wrapper to safely handle weights loading for older model formats.

### 3. How to Run (Standalone)

#### **Prerequisites**

* **Docker & Docker Compose** (Recommended due to heavy system dependencies like OpenCV and PyTorch).
* **Hardware:** At least 4GB of RAM (ML models are heavy). GPU support is optional but recommended for high-load production.
* **Disk Space:** ~2GB for Docker image and model weights.

#### **Quick Start with Docker**

The easiest way to start is using the provided Compose file:

```bash
docker-compose up -d

```

*The service will be available at `http://localhost:8000`. Models will be downloaded automatically on the first request and cached in a Docker volume.*

#### **Manual Run (Development)**

1. **Install system dependencies (Linux):**
```bash
sudo apt-get install libgl1 libglib2.0-0

```

2. **Install Python requirements:**
```bash
pip install -r requirements.txt

```

3. **Start the API:**
```bash
python main.py

```


#### **Testing the API**

You can test the service using `curl`:

```bash
curl -X POST http://localhost:8000/recognize \
  -F "file=@/path/to/your/truck_photo.jpg"

```