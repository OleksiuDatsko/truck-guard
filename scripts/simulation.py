
import requests
import time
import json
import os

# Configuration
BASE_URL = os.getenv("BASE_URL", "http://localhost")
AUTH_URL = f"{BASE_URL}/auth"
CORE_API_URL = f"{BASE_URL}/api"

ADMIN_USER = os.getenv("ADMIN_USER", "admin")
ADMIN_PASS = os.getenv("ADMIN_DEFAULT_PASSWORD", "secret123")

def get_admin_token():
    print(f"Logging in as {ADMIN_USER}...")
    try:
        resp = requests.post(f"{AUTH_URL}/login", json={
            "username": ADMIN_USER,
            "password": ADMIN_PASS
        })
        if resp.status_code == 200:
            token = resp.json().get("token")
            print("Login successful")
            return token
        else:
            print(f"Login failed: {resp.status_code} {resp.text}")
            return None
    except Exception as e:
        print(f"Login connection error: {e}")
        return None

def create_plate_event(token, camera_id, plate, timestamp=None):
    if timestamp is None:
        timestamp = time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())
    
    payload = {
        "camera_id": camera_id,
        "plate": plate,
        "timestamp": timestamp,
        "image_key": "simulated_image.jpg"
    }
    
    headers = {"Authorization": f"Bearer {token}"}
    response = requests.post(f"{CORE_API_URL}/events/plate", json=payload, headers=headers)
    print(f"Plate Event {plate} -> {response.status_code}")
    if response.status_code != 202:
        print(response.text)
    return response

def create_weight_event(token, scale_id, weight):
    timestamp = time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())
    payload = {
        "scale_id": scale_id,
        "weight": weight,
        "timestamp": timestamp
    }
    headers = {"Authorization": f"Bearer {token}"}
    response = requests.post(f"{CORE_API_URL}/events/weight", json=payload, headers=headers)
    print(f"Weight Event {weight} -> {response.status_code}")
    return response

def get_permits(token):
    headers = {"Authorization": f"Bearer {token}"}
    response = requests.get(f"{CORE_API_URL}/permits/", headers=headers)
    if response.status_code == 200:
        return response.json().get("data", [])
    print(f"Get Permits Failed: {response.text}")
    return []

def cleanup_infrastructure(token):
    headers = {"Authorization": f"Bearer {token}"}
    print("🧹 Cleaning up existing system state...")
    
    # 1. Clean up configs
    for ep in ["scales", "cameras", "excluded-plates"]:
        try:
            resp = requests.get(f"{CORE_API_URL}/configs/{ep}", headers=headers)
            if resp.status_code == 200:
                items = resp.json().get('data', []) if isinstance(resp.json(), dict) else resp.json()
                if items:
                    print(f"  Deleting {len(items)} {ep}...")
                    for i in items:
                        requests.delete(f"{CORE_API_URL}/configs/{ep}/{i['ID']}", headers=headers)
        except Exception as e:
            print(f"  Warning: failed to clean up {ep}: {e}")

    # 2. Clean up data entities
    for ep in ["posts", "modes", "vehicle-types", "payment-types", "companies"]:
        try:
            resp = requests.get(f"{CORE_API_URL}/data/{ep}", headers=headers)
            if resp.status_code == 200:
                items = resp.json().get('data', []) if isinstance(resp.json(), dict) else resp.json()
                if items:
                    print(f"  Deleting {len(items)} {ep}...")
                    for i in items:
                        requests.delete(f"{CORE_API_URL}/data/{ep}/{i['ID']}", headers=headers)
        except Exception as e:
            print(f"  Warning: failed to clean up {ep}: {e}")
    
    print("✨ Cleanup Complete")

def setup_infrastructure(token):
    headers = {"Authorization": f"Bearer {token}"}
    print("Setting up infrastructure...")
    
    # 1. Create Customs Post
    post_payload = {
        "name": "Simulation Post",
        "description": "Created by simulation.py"
    }
    
    resp = requests.post(f"{CORE_API_URL}/data/posts", json=post_payload, headers=headers)
    if resp.status_code in [200, 201]:
        post_id = resp.json()["ID"]
        print(f"Customs Post Created: {post_id}")
    else:
        print(f"Failed to create post: {resp.status_code} {resp.text}")
        return None

    # 2. Create Camera Front
    cam_front = {
        "name": "Front Cam",
        "type": "front",
        "customs_post_id": post_id,
        "trigger_permit_creation": True
    }
    resp = requests.post(f"{CORE_API_URL}/configs/cameras", json=cam_front, headers=headers)
    if resp.status_code not in [200, 201]:
        print(f"Failed to create front camera: {resp.status_code} {resp.text}")
        return None
    cam_front_id = resp.json()["camera"]["camera_id"]

    # 3. Create Camera Back
    cam_back = {
        "name": "Back Cam",
        "type": "back",
        "format": "json",
        "customs_post_id": post_id
    }
    resp = requests.post(f"{CORE_API_URL}/configs/cameras", json=cam_back, headers=headers)
    if resp.status_code not in [200, 201]:
        print(f"Failed to create back camera: {resp.status_code} {resp.text}")
        return None
    cam_back_id = resp.json()["camera"]["camera_id"]

    # 4. Create Scale
    scale = {
        "name": "Main Scale",
        "customs_post_id": post_id,
        "trigger_permit_creation": True
    }
    resp = requests.post(f"{CORE_API_URL}/configs/scales", json=scale, headers=headers)
    if resp.status_code not in [200, 201]:
        print(f"Failed to create scale: {resp.status_code} {resp.text}")
        return None
    scale_id = resp.json()["scale"]["scale_id"]
    
    print("Infrastructure Setup Complete")
    return {
        "post_id": post_id,
        "cam_front_id": cam_front_id,
        "cam_back_id": cam_back_id,
        "scale_id": scale_id
    }

def main():
    print("Waiting for service to be healthy...")
    for _ in range(10):
        try:
            if requests.get(f"{AUTH_URL}/health").status_code == 200:
                break
        except:
            time.sleep(1)
    
    token = get_admin_token()
    if not token:
        print("Aborting due to login failure")
        return

    cleanup_infrastructure(token)

    infra = setup_infrastructure(token)
    if not infra:
        print("Aborting simulation due to setup failure")
        return

    print("Starting Simulation...")

    plate = f"SIM{int(time.time())%10000}"
    
    print(f"--- Simulating Front Camera Scan {plate} ---")
    create_plate_event(token, infra["cam_front_id"], plate)
    
    time.sleep(2)
    
    permits = get_permits(token)
    if not permits:
        print("FAIL: No permits found")
        # Check if they are matched asynchronously
        time.sleep(2)
        permits = get_permits(token)
        if not permits:
            return
    
    # Permit matching might take a second
    latest_permit = permits[0]
    print(f"Latest Permit ID: {latest_permit['ID']}, Plate Front: {latest_permit['plate_front']}")
    
    # 2. Back Camera Scan
    print("--- Simulating Back Camera Scan ---")
    create_plate_event(token, infra["cam_back_id"], plate)
    
    time.sleep(1)
    
    # 3. Weight Scan
    print("--- Simulating Weight Scan ---")
    create_weight_event(token, infra["scale_id"], 20000)
    
    # Verify
    print("--- Verification ---")
    time.sleep(2)
    permits = get_permits(token)
    for p in permits:
        if p['plate_front'] == plate:
            print(f"Permit {p['ID']} found!")
            print(f"  Plate Back: {p['plate_back']}")
            print(f"  Weight: {p['total_weight']}")
            if p['plate_back'] == plate and p['total_weight'] > 0:
                print("✅ Simulation Success!")
            else:
                print("⚠️ Simulation Data Incomplete")
            return

    print("❌ Permit not found in verification step")

if __name__ == "__main__":
    main()
