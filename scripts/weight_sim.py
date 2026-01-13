import requests
import json
import time
import random
import os

# Configuration
BASE_URL = os.getenv("BASE_URL", "http://localhost")
AUTH_URL = f"{BASE_URL}/auth"
CORE_API_URL = f"{BASE_URL}/api"
INGEST_URL = f"{BASE_URL}/ingest/weight"

ADMIN_USER = os.getenv("ADMIN_USER", "admin")
ADMIN_PASS = os.getenv("ADMIN_DEFAULT_PASSWORD", "admin123")

# Weight Scenarios
WEIGHT_SCENARIOS = [
    {
        "id": "SCALE_01",
        "name": "North Gate Scale",
        "description": "Main weighing station at the north gate",
        "template": lambda weight: json.dumps({
            "weight": weight,
            "unit": "kg",
            "confidence": 1.0
        })
    },
    {
        "id": "SCALE_02",
        "name": "South Gate Scale",
        "description": "Exit weighing station",
        "template": lambda weight: json.dumps({
            "weight": weight,
            "unit": "kg",
            "confidence": 1.0
        })
    }
]

def get_admin_token():
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

def setup_scale(token, scenario):
    headers = {"Authorization": f"Bearer {token}"}
    scale_name = scenario["name"]
    
    try:
        # Check if exists
        resp = requests.get(f"{CORE_API_URL}/configs/scales", headers=headers, timeout=5)
        
        cameras = resp.json().get("data", [])
        for cam in cameras:
            if cam.get("name") == scale_name:
                cam_id = cam.get("ID") or cam.get("id")
                requests.delete(f"{CORE_API_URL}/configs/scales/{cam_id}", headers=headers, timeout=5)
                print(f"🗑️  Cleaned up existing scale: {scale_name}")

        payload = {
            "name": scale_name,
            "description": scenario.get("description", ""),
            "format": "json",
            "field_mapping": "{\"weight\": \"weight\"}"
        }
        
        resp = requests.post(f"{CORE_API_URL}/configs/scales", json=payload, headers=headers, timeout=5)
        if resp.status_code == 201:
            data = resp.json()
            api_key = data.get("api_key")
            print(f"✅ Scale '{scale_name}' registered. Key obtained.")
            return api_key
        else:
            print(f"❌ Failed to create scale: {resp.status_code} {resp.text}")
            return None
            
    except Exception as e:
        print(f"🚨 Error during scale setup: {e}")
        return None

def send_weight_event(scenario, api_key):
    weight = round(random.uniform(5000, 40000), 2)
    payload = scenario["template"](weight)

    data = {
        'device_id': scenario["id"],
        'payload': payload
    }
    headers = {'X-API-Key': api_key}

    try:
        print(f"⚖️ [{scenario['name']}] Sending {weight} kg...")
        resp = requests.post(INGEST_URL, data=data, headers=headers, timeout=5)
        
        if resp.status_code == 202:
            print(f"  ✅ Accepted (202)")
        else:
            print(f"  ❌ Failed ({resp.status_code}): {resp.text}")
    except Exception as e:
        print(f"  🚨 Connection error: {e}")

def main():
    print("🚀 Starting Weight Simulator...")
    
    token = get_admin_token()
    if not token:
        return

    for scenario in WEIGHT_SCENARIOS:
        api_key = setup_scale(token, scenario)
        if api_key:
            scenario["api_key"] = api_key

    active_scenarios = [s for s in WEIGHT_SCENARIOS if "api_key" in s]
    
    if not active_scenarios:
        print("Aborting: No active scale scenarios.")
        return

    try:
        while True:
            current_scale = random.choice(active_scenarios)
            send_weight_event(current_scale, current_scale["api_key"])
            time.sleep(random.randint(5, 10))
    except KeyboardInterrupt:
        print("\n🛑 Simulator stopped.")

if __name__ == "__main__":
    main()
