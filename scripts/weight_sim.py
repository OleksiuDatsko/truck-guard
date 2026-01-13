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

def setup_scale(token, scenario, gate_id=None):
    headers = {"Authorization": f"Bearer {token}"}
    scale_name = scenario["name"]
    
    try:
        # Check if exists
        resp = requests.get(f"{CORE_API_URL}/configs/scales", headers=headers, timeout=5)
        if resp.status_code == 200:
            scales = resp.json().get("data", [])
            for scale in scales:
                if scale.get("name") == scale_name:
                    scale_id = scale.get("ID") or scale.get("id")
                    requests.delete(f"{CORE_API_URL}/configs/scales/{scale_id}", headers=headers, timeout=5)
                    print(f"🗑️  Cleaned up existing scale: {scale_name}")

        payload = {
            "name": scale_name,
            "description": scenario.get("description", ""),
            "format": "json",
            "field_mapping": json.dumps({"weight": "weight"}),
            "gate_id": gate_id
        }
        
        resp = requests.post(f"{CORE_API_URL}/configs/scales", json=payload, headers=headers, timeout=5)
        if resp.status_code == 201:
            data = resp.json()
            api_key = data.get("api_key")
            print(f"✅ Scale '{scale_name}' registered at Gate {gate_id}. Key obtained.")
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

    # Setup Gate
    gate_id = get_or_create_gate(token)

    for scenario in WEIGHT_SCENARIOS:
        api_key = setup_scale(token, scenario, gate_id)
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
