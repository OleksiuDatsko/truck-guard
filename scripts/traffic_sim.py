import requests
import json
import time
import random
import io
import os
from PIL import Image

# Configuration
BASE_URL = os.getenv("BASE_URL", "http://localhost")
AUTH_URL = f"{BASE_URL}/auth"
CORE_API_URL = f"{BASE_URL}/api"
INGEST_CAMERA_URL = f"{BASE_URL}/ingest/camera"
INGEST_WEIGHT_URL = f"{BASE_URL}/ingest/weight"

ADMIN_USER = os.getenv("ADMIN_USER", "admin")
ADMIN_PASS = os.getenv("ADMIN_DEFAULT_PASSWORD", "admin123")

def get_admin_token():
    resp = requests.post(f"{AUTH_URL}/login", json={"username": ADMIN_USER, "password": ADMIN_PASS})
    print(resp)
    return resp.json().get("token")

def setup_gate(token):
    headers = {"Authorization": f"Bearer {token}"}
    gate_name = f"Unified Sim Gate {random.randint(100, 999)}"
    resp = requests.post(f"{CORE_API_URL}/configs/gates", json={"name": gate_name}, headers=headers)
    gate = resp.json()
    return gate["ID"], gate["name"]

def setup_device(token, type, name, gate_id, device_id):
    headers = {"Authorization": f"Bearer {token}"}
    endpoint = "cameras" if type == "camera" else "scales"
    
    payload = {
        "name": name,
        "format": "json",
        "gate_id": gate_id,
        "field_mapping": json.dumps({"plate": "plate"}) if type == "camera" else json.dumps({"weight": "weight"})
    }
    
    if type == "camera":
        payload["camera_id"] = device_id
    else:
        payload["scale_id"] = device_id

    resp = requests.post(f"{CORE_API_URL}/configs/{endpoint}", json=payload, headers=headers)
    return resp.json().get("api_key")

def generate_image():
    file = io.BytesIO()
    color = (random.randint(0, 255), random.randint(0, 255), random.randint(0, 255))
    Image.new('RGB', (640, 480), color=color).save(file, 'jpeg')
    file.seek(0)
    return file

def send_plate(api_key, device_id, plate):
    files = {'image': ('frame.jpg', generate_image(), 'image/jpeg')}
    data = {'device_id': device_id, 'payload': json.dumps({"plate": plate})}
    requests.post(INGEST_CAMERA_URL, files=files, data=data, headers={'X-API-Key': api_key})

def send_weight(api_key, device_id, weight):
    data = {'device_id': device_id, 'payload': json.dumps({"weight": weight})}
    requests.post(INGEST_WEIGHT_URL, data=data, headers={'X-API-Key': api_key})

def main():
    print("🚀 Starting Unified Traffic Simulator...")
    token = get_admin_token() or exit(1)
    
    # 1. Setup Infrastructure
    gate_id, gate_name = setup_gate(token)
    print(f"🏢 Created Gate: {gate_name}")
    
    # Entrance Cameras
    cam1_key = setup_device(token, "camera", "Entrance Front", gate_id, "CAM_ENTRANCE_FRONT")
    cam2_key = setup_device(token, "camera", "Entrance Back", gate_id, "CAM_ENTRANCE_BACK")
    
    # Scale Cameras
    cam3_key = setup_device(token, "camera", "Scale Front", gate_id, "CAM_SCALE_FRONT")
    cam4_key = setup_device(token, "camera", "Scale Back", gate_id, "CAM_SCALE_BACK")
    
    # Scale
    scale_key = setup_device(token, "scale", "Main Scale", gate_id, "SCALE_MAIN")
    
    print("✅ Infrastructure Ready. Simulating Traffic...")
    
    while True:
        plate_front = f"AA{random.randint(1000, 9999)}BB"
        plate_back = f"AA{random.randint(1000, 9999)}BB"
        weight = round(random.uniform(5000, 30000), 2)
        
        print(f"\n🚛 Truck {plate_front} Arriving...")
        
        # 1. Entrance (2 events)
        send_plate(cam1_key, "CAM_ENTRANCE_FRONT", plate_front)
        print("  📸 Entrance Front")
        time.sleep(0.5)
        send_plate(cam2_key, "CAM_ENTRANCE_BACK", plate_back)
        print("  📸 Entrance Back")
        
        # Drive to scale
        delay = random.randint(2, 5)
        print(f"  ... driving to scale ({delay}s) ...")
        time.sleep(delay)
        
        # 2. Scale Cameras (2 events)
        send_plate(cam3_key, "CAM_SCALE_FRONT", plate_front)
        print("  📸 Scale Front")
        time.sleep(0.5)
        send_plate(cam4_key, "CAM_SCALE_BACK", plate_back)
        print("  📸 Scale Back")
        
        
        # 3. Weighing (1 event) - The 'Close' event
        time.sleep(1)
        send_weight(scale_key, "SCALE_MAIN", weight)
        print(f"  ⚖️  Weighed: {weight}kg")
        
        print(f"✅ Truck {plate_front} Processed.")
        
        time.sleep(random.randint(5, 10))

if __name__ == "__main__":
    main()
