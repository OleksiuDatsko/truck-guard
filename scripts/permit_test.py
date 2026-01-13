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
    return resp.json().get("token")

def setup_gate(token):
    headers = {"Authorization": f"Bearer {token}"}
    gate_name = f"Test Gate {random.randint(100, 999)}"
    resp = requests.post(f"{CORE_API_URL}/configs/gates", json={"name": gate_name}, headers=headers)
    gate = resp.json()
    return gate["ID"], gate["name"]

def setup_device(token, type, name, gate_id):
    headers = {"Authorization": f"Bearer {token}"}
    endpoint = "cameras" if type == "camera" else "scales"
    payload = {
        "name": name,
        "format": "json",
        "gate_id": gate_id,
        "camera_id": "CAM_TEST" if type == "camera" else None,
        "scale_id": "SCALE_TEST" if type == "scale" else None,
        "field_mapping": json.dumps({"plate": "plate"}) if type == "camera" else json.dumps({"weight": "weight"})
    }
    resp = requests.post(f"{CORE_API_URL}/configs/{endpoint}", json=payload, headers=headers)
    return resp.json().get("api_key")

def generate_image():
    file = io.BytesIO()
    Image.new('RGB', (100, 100), color='blue').save(file, 'jpeg')
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
    print("🧪 Starting End-to-End Permit Correlation Test...")
    token = get_admin_token() or exit(1)
    
    gate_id, gate_name = setup_gate(token)
    print(f"🏢 Created {gate_name}")
    
    cam_key = setup_device(token, "camera", f"Cam {gate_name}", gate_id)
    scale_key = setup_device(token, "scale", f"Scale {gate_name}", gate_id)
    
    target_plate_front = f"TEST{random.randint(1000, 9999)}"
    target_plate_back = f"TRAIL{random.randint(1000, 9999)}"
    target_weight = 25500.0
    
    print(f"🚛 Simulating passage: {target_plate_front} + {target_plate_back} with {target_weight}kg")
    
    # 1. Front Plate
    send_plate(cam_key, "CAM_TEST", target_plate_front)
    time.sleep(1)
    
    # 2. Back Plate
    send_plate(cam_key, "CAM_TEST", target_plate_back)
    time.sleep(1)
    
    # 3. Weight
    send_weight(scale_key, "SCALE_TEST", target_weight)
    
    print("⏳ Waiting for processing...")
    time.sleep(3)
    
    # 4. Verify
    headers = {"Authorization": f"Bearer {token}"}
    resp = requests.get(f"{CORE_API_URL}/permits", headers=headers)
    
    if resp.status_code != 200:
        print(f"❌ Error fetching permits: Status {resp.status_code}")
        print(f"Response Body: {resp.text[:500]}")
        return

    try:
        permits = resp.json().get("data", [])
    except json.JSONDecodeError:
        print(f"❌ Failed to decode JSON response. Status: {resp.status_code}")
        print(f"Response Body: {resp.text[:500]}")
        return
    
    found = False
    for p in permits:
        if p.get("plate_front") == target_plate_front and p.get("plate_back") == target_plate_back:
            # Check for linked events
            plate_events = p.get("plate_events", [])
            weight_events = p.get("weight_events", [])
            
            if len(plate_events) >= 2 and len(weight_events) >= 1:
                if abs(p.get("total_weight", 0) - target_weight) < 1:
                    print(f"✅ Success! Permit found with {len(plate_events)} plate events and {len(weight_events)} weight events.")
                    print(f"  📸 Images linked via events: {[e.get('image_key') for e in plate_events]}")
                    found = True
                    break
    
    if not found:
        print("❌ Test failed: Correlated permit not found or data mismatch.")
        print("Latest permits:", json.dumps(permits[:2], indent=2))

if __name__ == "__main__":
    main()
