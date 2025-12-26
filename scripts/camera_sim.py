import requests
import json
import time
import random
import io
from PIL import Image

# Налаштування
INGESTOR_URL = "http://localhost/ingest"
DEVICE_ID = "CAM_Lviv_01"
# API ключ, якщо ви додали перевірку в Ingestor (X-API-Key)
API_KEY = "2ee2c68870715e719922de095681a018" 

def generate_dummy_image():
    """Створює просте кольорове зображення в пам'яті."""
    file = io.BytesIO()
    # Випадковий колір фону для візуальної різниці
    color = (random.randint(0, 255), random.randint(0, 255), random.randint(0, 255))
    image = Image.new('RGB', (640, 480), color=color)
    image.save(file, 'jpeg')
    file.seek(0)
    return file

def simulate_event():
    # Симулюємо дані, які зазвичай шле камера (наприклад, Hikvision)
    payload = {
        "timestamp": int(time.time()),
        "plate_number": f"BC{random.randint(1000, 9999)}AX",
        "confidence": round(random.uniform(0.75, 0.99), 2),
        "location": "Checkpoint-1"
    }

    image_file = generate_dummy_image()

    # Підготовка multipart/form-data
    files = {
        'image': ('camera_frame.jpg', image_file, 'image/jpeg')
    }
    data = {
        'device_id': DEVICE_ID,
        'payload': json.dumps(payload)
    }
    headers = {
        'X-API-Key': API_KEY
    }

    try:
        print(f"🚀 Sending event for {payload['plate_number']}...")
        response = requests.post(INGESTOR_URL, files=files, data=data, headers=headers)
        
        if response.status_code == 202:
            print(f"✅ Accepted: {response.status_code}")
        else:
            print(f"❌ Failed: {response.status_code} - {response.text}")
            
    except Exception as e:
        print(f"🚨 Connection error: {e}")

if __name__ == "__main__":
    print("📸 Camera Simulator started. Press Ctrl+C to stop.")
    while True:
        simulate_event()
        # Пауза між "проїздами" фур (від 2 до 5 секунд)
        sleep_time = random.randint(2, 5)
        time.sleep(sleep_time)