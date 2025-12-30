import requests
import json
import time
import random
import io
from PIL import Image

# Налаштування
INGESTOR_URL = "http://localhost/ingest"

# Конфігурації різних типів камер
CAMERA_SCENARIOS = [
    {
        "id": "CAM_JSON_01",
        "name": "Lviv Entrance (JSON)",
        "api_key": "9c4299e1eb796677ae85192ced8e3a3c",
        "format": "json",
        "template": lambda plate: json.dumps({
            "event_type": "plate_recognition",
            "data": {
                "plate_number": plate,
                "confidence": round(random.uniform(0.8, 0.99), 2)
            },
            "metadata": {"location": "A1-Entrance"}
        })
    },
    {
        "id": "CAM_XML_02",
        "name": "Kyiv Highway (XML)",
        "api_key": "1e8d1fa6a8e21cc1c9b325f6b4ec2100",
        "format": "xml",
        "template": lambda plate: f"""
        <Event>
            <DeviceID>XML_CAM_02</DeviceID>
            <Vehicle>
                <Plate>{plate}</Plate>
                <Speed>{random.randint(40, 90)}</Speed>
            </Vehicle>
            <Timestamp>{int(time.time())}</Timestamp>
        </Event>
        """
    }
]

def generate_image():
    """Створює випадкове зображення."""
    file = io.BytesIO()
    color = (random.randint(0, 255), random.randint(0, 255), random.randint(0, 255))
    image = Image.new('RGB', (800, 600), color=color)
    image.save(file, 'jpeg')
    file.seek(0)
    return file

def send_camera_event(scenario):
    """Генерує номер та надсилає multipart запит."""
    plate = f"BC{random.randint(1000, 9999)}HX"
    payload = scenario["template"](plate)
    image = generate_image()

    files = {'image': ('frame.jpg', image, 'image/jpeg')}
    data = {
        'device_id': scenario["id"],
        'payload': payload
    }
    headers = {'X-API-Key': scenario["api_key"]}

    try:
        print(f"📸 [{scenario['name']}] Sending {plate} in {scenario['format']}...")
        resp = requests.post(INGESTOR_URL, files=files, data=data, headers=headers, timeout=5)
        
        if resp.status_code == 202:
            print(f"  ✅ Accepted (202)")
        else:
            print(f"  ❌ Failed ({resp.status_code}): {resp.text}")
    except Exception as e:
        print(f"  🚨 Connection error: {e}")

if __name__ == "__main__":
    print("🚀 Starting Multi-Camera Simulator...")
    print("Ensure you have created these cameras in Core API first!")
    
    try:
        while True:
            # Вибираємо випадкову камеру для симуляції події
            current_camera = random.choice(CAMERA_SCENARIOS)
            send_camera_event(current_camera)
            
            # Пауза між подіями
            # time.sleep(random.randint(3, 7))
    except KeyboardInterrupt:
        print("\n🛑 Simulator stopped.")