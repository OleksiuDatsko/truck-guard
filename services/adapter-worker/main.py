import os
import json
import time
import requests
import xmltodict
from redis import Redis
from minio import Minio
from datetime import datetime

CORE_URL = os.getenv("CORE_URL", "http://gateway/api/v1")
WORKER_API_KEY = os.getenv("WORKER_API_KEY", "worker_internal_secret_2025_token")
ANPR_URL = os.getenv("ANPR_URL", "http://anpr:8000/recognize")
REDIS_ADDR = os.getenv("REDIS_ADDR", "redis:6379")
MINIO_BUCKET = os.getenv("MINIO_BUCKET", "truck-images")

redis_client = Redis.from_url(f"redis://{REDIS_ADDR}", decode_responses=True)
minio_client = Minio(
    os.getenv("MINIO_ENDPOINT", "minio:9000"),
    access_key=os.getenv("MINIO_ACCESS_KEY", "minioadmin"),
    secret_key=os.getenv("MINIO_SECRET_KEY", "minioadmin"),
    secure=False
)

config_cache = {}
CACHE_TTL = 60 

def log(message, level="INFO"):
    """Уніфікована функція логування (flush=True для Docker)"""
    timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    print(f"[{timestamp}] [{level}] {message}", flush=True)

def get_auth_headers():
    """Заголовки авторизації для Gateway [cite: 32]"""
    return {
        "X-API-Key": WORKER_API_KEY,
        "Content-Type": "application/json"
    }

def get_camera_config(camera_id):
    now = time.time()
    if camera_id in config_cache:
        cached_data, timestamp = config_cache[camera_id]
        if now - timestamp < CACHE_TTL:
            return cached_data

    try:
        url = f"{CORE_URL}/cameras/by-id/{camera_id}"
        log(f"Fetching config: {url}")
        
        resp = requests.get(url, headers=get_auth_headers(), timeout=5)
        
        if resp.status_code == 200:
            config = resp.json()
            if isinstance(config.get('field_mapping'), str) and config['field_mapping']:
                config['field_mapping'] = json.loads(config['field_mapping'])
            config_cache[camera_id] = (config, now)
            log(f"Config loaded for camera {camera_id}")
            return config
        else:
            log(f"Failed to get config. Status: {resp.status_code}, Body: {resp.text}", "ERROR")
    except Exception as e:
        log(f"Request error (get_config): {e}", "ERROR")
    
    return None

def process_message(msg_id, redis_data):
    try:
        raw_json = redis_data.get("data")
        if not raw_json: return
        
        event_data = json.loads(raw_json)
        camera_id = event_data.get("camera_id")
        image_key = event_data.get("image_key")
        raw_payload = event_data.get("payload")

        log(f"===> Processing message {msg_id} (Camera: {camera_id})")

        config = get_camera_config(camera_id)
        if not config:
            log(f"Skipping message {msg_id}: config not available", "WARNING")
            return

        parsed_body = xmltodict.parse(raw_payload) if config.get("format") == "xml" else json.loads(raw_payload)

        mapping = config.get("field_mapping") or {}
        plate = None
        path = mapping.get("plate")
        
        if path:
            keys = path.split('/')
            temp_data = parsed_body
            for k in keys:
                if isinstance(temp_data, dict): temp_data = temp_data.get(k)
            plate = temp_data

        if not plate or config.get("run_anpr"):
            log(f"Running ANPR for {image_key}...")
            img_obj = minio_client.get_object(MINIO_BUCKET, image_key)
            anpr_resp = requests.post(ANPR_URL, files={"file": img_obj.read()}, timeout=10)
            if anpr_resp.status_code == 200:
                plates = anpr_resp.json().get("plates", [])
                if plates: plate = plates[0].get("plate")
            else:
                log(f"ANPR service error: {anpr_resp.status_code}", "ERROR")

        if plate:
            final_event = {
                "camera_id": camera_id,
                "camera_name": config.get("name", camera_id),
                "plate": plate.upper().replace(" ", ""),
                "image_key": image_key,
                "timestamp": event_data.get("at")
            }
            
            url = f"{CORE_URL}/events/plate"
            log(f"Sending event to Core: {url} | Plate: {plate}")
            
            resp = requests.post(url, json=final_event, headers=get_auth_headers(), timeout=5)
            
            if resp.status_code in [200, 201, 202]:
                log(f"✅ SUCCESSFULLY recorded: {plate}")
            else:
                log(f"❌ CORE REJECTED: {resp.status_code} | Body: {resp.text}", "ERROR")
        else:
            log(f"No plate found for message {msg_id}", "WARNING")

    except Exception as e:
        log(f"Processing error: {e}", "CRITICAL")

def main():
    log(f"Worker starting. Redis: {REDIS_ADDR}, Core: {CORE_URL}")
    last_id = "0" 
    while True:
        try:
            streams = redis_client.xread({"camera:raw": last_id}, count=1, block=5000)
            if streams:
                for stream, messages in streams:
                    for msg_id, data in messages:
                        process_message(msg_id, data)
                        last_id = msg_id
        except Exception as e:
            log(f"Redis connection error: {e}", "ERROR")
            time.sleep(2)

if __name__ == "__main__":
    main()