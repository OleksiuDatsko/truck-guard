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
INGEST_URL = f"{BASE_URL}/ingest/camera"

ADMIN_USER = os.getenv("ADMIN_USER", "admin")
ADMIN_PASS = os.getenv("ADMIN_DEFAULT_PASSWORD", "admin123")

# Camera Scenarios
CAMERA_SCENARIOS = [
    {
        "id": "CAM_JSON_01",
        "name": "Lviv Entrance (JSON)",
        "format": "json",
        "description": "Main entrance simulation with JSON payload",
        "field_mapping": {"plate": "data/plate_number"},
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
        "format": "xml",
        "description": "Highway speed monitoring simulation with XML payload",
        "field_mapping": {"plate": "Event/Vehicle/Plate"},
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

def get_admin_token():
    """Logs in as admin and returns the JWT token."""
    try:
        resp = requests.post(f"{AUTH_URL}/login", json={
            "username": ADMIN_USER,
            "password": ADMIN_PASS
        }, timeout=5)
        if resp.status_code == 200:
            return resp.json().get("token")
        else:
            print(f"❌ Login failed: {resp.status_code} {resp.text}")
            return None
    except Exception as e:
        print(f"🚨 Connection error during login: {e}")
        return None

def get_or_create_gate(token, gate_name="Main Entrance"):
    """Finds or creates a Gate in the Core API."""
    headers = {"Authorization": f"Bearer {token}"}
    try:
        resp = requests.get(f"{CORE_API_URL}/configs/gates", headers=headers, timeout=5)
        if resp.status_code == 200:
            gates = resp.json()
            for gate in gates:
                if gate.get("name") == gate_name:
                    print(f"📍 Found existing Gate: {gate_name} (ID: {gate['ID']})")
                    return gate["ID"]
        
        # Create if not found
        payload = {"name": gate_name, "description": "Auto-created by simulator"}
        resp = requests.post(f"{CORE_API_URL}/configs/gates", json=payload, headers=headers, timeout=5)
        if resp.status_code == 201:
            gate = resp.json()
            print(f"🆕 Created new Gate: {gate_name} (ID: {gate['ID']})")
            return gate["ID"]
        else:
            print(f"❌ Failed to create Gate: {resp.status_code} {resp.text}")
            return None
    except Exception as e:
        print(f"🚨 Error during gate setup: {e}")
        return None

def setup_camera(token, scenario, gate_id=None):
    """
    Ensures camera exists in Core API and links it to a gate. 
    Returns the API Key for ingestion.
    """
    headers = {"Authorization": f"Bearer {token}"}
    camera_name = scenario["name"]
    
    try:
        # 1. Clean up existing camera with same name to get a fresh API Key
        resp = requests.get(f"{CORE_API_URL}/configs/cameras", headers=headers, timeout=5)
        if resp.status_code == 200:
            cameras = resp.json()
            for cam in cameras.get("data", []):
                if cam.get("name") == camera_name:
                    cam_id = cam.get("ID") or cam.get("id")
                    if cam_id:
                        requests.delete(f"{CORE_API_URL}/configs/cameras/{cam_id}", headers=headers, timeout=5)
                        print(f"🗑️  Cleaned up existing camera: {camera_name}")
                    break
        
        # 2. Create camera config in Core
        payload = {
            "name": camera_name,
            "description": scenario.get("description", ""),
            "format": scenario["format"],
            "field_mapping": json.dumps(scenario.get("field_mapping", {})),
            "gate_id": gate_id
        }
        
        resp = requests.post(f"{CORE_API_URL}/configs/cameras", json=payload, headers=headers, timeout=5)
        if resp.status_code == 201:
            data = resp.json()
            api_key = data.get("api_key")
            print(f"✅ Camera '{camera_name}' registered at Gate {gate_id}. Key obtained.")
            return api_key
        else:
            print(f"❌ Failed to create camera: {resp.status_code} {resp.text}")
            return None
            
    except Exception as e:
        print(f"🚨 Error during camera setup: {e}")
        return None

def generate_image():
    """Creates a random image."""
    file = io.BytesIO()
    color = (random.randint(0, 255), random.randint(0, 255), random.randint(0, 255))
    image = Image.new('RGB', (800, 600), color=color)
    image.save(file, 'jpeg')
    file.seek(0)
    return file

def send_camera_event(scenario, api_key):
    """Generates plate and sends multipart request."""
    plate = f"BC{random.randint(1000, 9999)}HX"
    payload = scenario["template"](plate)
    image = generate_image()

    files = {'image': ('frame.jpg', image, 'image/jpeg')}
    data = {
        'device_id': scenario["id"],
        'payload': payload
    }
    headers = {'X-API-Key': api_key}

    try:
        print(f"📸 [{scenario['name']}] Sending {plate}...")
        resp = requests.post(INGEST_URL, files=files, data=data, headers=headers, timeout=5)
        
        if resp.status_code == 202:
            print(f"  ✅ Accepted (202)")
        else:
            print(f"  ❌ Failed ({resp.status_code}): {resp.text}")
    except Exception as e:
        print(f"  🚨 Connection error: {e}")

def main():
    print("🚀 Starting Autonomous Multi-Camera Simulator...")
    
    token = get_admin_token()
    if not token:
        print("Aborting: Could not obtain admin token.")
        return

    # Setup Gate
    gate_id = get_or_create_gate(token)

    # Setup all cameras
    for scenario in CAMERA_SCENARIOS:
        api_key = setup_camera(token, scenario, gate_id)
        if api_key:
            scenario["api_key"] = api_key
        else:
            print(f"⚠️  Skipping scenario {scenario['name']} due to setup failure.")

    # Filter active scenarios
    active_scenarios = [s for s in CAMERA_SCENARIOS if "api_key" in s]
    
    if not active_scenarios:
        print("Aborting: No active camera scenarios.")
        return

    print(f"🔥 Simulation running for {len(active_scenarios)} cameras.")
    
    try:
        while True:
            current_camera = random.choice(active_scenarios)
            send_camera_event(current_camera, current_camera["api_key"])
            time.sleep(random.randint(3, 7))
    except KeyboardInterrupt:
        print("\n🛑 Simulator stopped.")

if __name__ == "__main__":
    main()