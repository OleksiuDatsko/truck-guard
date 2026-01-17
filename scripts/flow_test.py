import requests
import json
import time
import io
import os
import random
from PIL import Image
from datetime import datetime

# Конфігурація
BASE_URL = os.getenv("BASE_URL", "http://localhost")
AUTH_URL = f"{BASE_URL}/auth"
CORE_API_URL = f"{BASE_URL}/api"
INGEST_CAMERA_URL = f"{BASE_URL}/ingest/camera"
INGEST_WEIGHT_URL = f"{BASE_URL}/ingest/weight"

ADMIN_USER = os.getenv("ADMIN_USER", "admin")
ADMIN_PASS = os.getenv("ADMIN_DEFAULT_PASSWORD", "secret123")

# Дані фури
TRUCK = {"f": "BC7777EX", "b": "BC7777EE", "w": 32500}

def get_headers(token):
    return {"Authorization": f"Bearer {token}"}

def cleanup(token):
    """Видаляємо всі старі налаштування"""
    print("🧹 Повна очистка конфігурацій (Flows, Scales, Cameras, Gates)...")
    h = get_headers(token)
    for ep in ["flows", "scales", "cameras", "gates"]:
        try:
            resp = requests.get(f"{CORE_API_URL}/configs/{ep}", headers=h)
            items = resp.json().get('data', []) if isinstance(resp.json(), dict) else resp.json()
            if items:
                for i in items:
                    requests.delete(f"{CORE_API_URL}/configs/{ep}/{i['ID']}", headers=h)
        except: pass
    print("✨ Система чиста.")

def setup_env(token):
    """Налаштування: 3 гейти, по 2 камери на кожному"""
    h = get_headers(token)
    print("🏗️ Створення інфраструктури (по 2 камери на гейт)...")
    
    # 1. Гейти
    g_in = requests.post(f"{CORE_API_URL}/configs/gates", headers=h, json={"name":"ENTRY", "is_entry":True}).json()['ID']
    g_sc = requests.post(f"{CORE_API_URL}/configs/gates", headers=h, json={"name":"SCALE"}).json()['ID']
    g_out = requests.post(f"{CORE_API_URL}/configs/gates", headers=h, json={"name":"EXIT", "is_exit":True}).json()['ID']
    
    env_keys = {
        "gate_ids": [g_in, g_sc, g_out],
        "cam_keys": {}
    }

    # 2. Створюємо по 2 камери для кожного гейту
    gate_configs = [
        ("IN", g_in),
        ("SC", g_sc),
        ("OUT", g_out)
    ]

    for prefix, g_id in gate_configs:
        # Передня камера
        key_f = requests.post(f"{CORE_API_URL}/configs/cameras", headers=h, 
                              json={"name": f"{prefix}_Front", "gate_id": g_id, "format": "json", "field_mapping": '{"plate":"plate"}'}).json()['api_key']
        # Задня камера
        key_b = requests.post(f"{CORE_API_URL}/configs/cameras", headers=h, 
                              json={"name": f"{prefix}_Back", "gate_id": g_id, "format": "json", "field_mapping": '{"plate":"plate"}'}).json()['api_key']
        
        env_keys["cam_keys"][prefix] = [key_f, key_b]

    # 3. Вага
    s_key = requests.post(f"{CORE_API_URL}/configs/scales", headers=h, 
                          json={"name": "Main_Scale", "gate_id": g_sc, "format": "json", "field_mapping": '{"weight":"weight"}'}).json()['api_key']
    env_keys["scale_key"] = s_key

    return env_keys

def setup_flow(token, gate_ids):
    print("🌊 Налаштування Flow маршруту...")
    h = get_headers(token)
    requests.post(f"{CORE_API_URL}/configs/flows", headers=h, json={
        "name": "Повний цикл (2 камери)",
        "steps": [
            {"gate_id": gate_ids[0], "sequence": 1},
            {"gate_id": gate_ids[1], "sequence": 2},
            {"gate_id": gate_ids[2], "sequence": 3}
        ]
    })

def send_cam(key, plate, cam_label=""):
    f = io.BytesIO()
    Image.new('RGB', (100, 100), color=(random.randint(0,255), 50, 50)).save(f, 'jpeg')
    f.seek(0)
    requests.post(INGEST_CAMERA_URL, headers={'X-API-Key':key}, files={'image':('p.jpg',f)}, 
                  data={'device_id':'SIM','payload':json.dumps({"plate":plate})})
    print(f"   📸 {cam_label}: {plate}")

def send_weight(key, val):
    requests.post(INGEST_WEIGHT_URL, headers={'X-API-Key':key}, 
                  data={'device_id':'SCALE','payload':json.dumps({"weight":val})})
    print(f"   ⚖️  Вага: {val} kg")

def main():
    token = requests.post(f"{AUTH_URL}/login", json={"username":ADMIN_USER, "password":ADMIN_PASS}).json().get("token")
    if not token: return

    cleanup(token)
    env = setup_env(token)
    setup_flow(token, env['gate_ids'])
    
    k = env['cam_keys']
    s_k = env['scale_key']

    print("\n--- 🚛 ЕТАП 1: ЗАЇЗД (2 камери) ---")
    send_cam(k['IN'][0], TRUCK['f'], "ENTRY Front")
    time.sleep(0.5)
    send_cam(k['IN'][1], TRUCK['b'], "ENTRY Back")
    
    print("\n--- ⚖️  ЕТАП 2: ВАГА (2 камери + вага) ---")
    send_cam(k['SC'][0], TRUCK['f'], "SCALE Front")
    send_cam(k['SC'][1], TRUCK['b'], "SCALE Back")
    time.sleep(1) 
    send_weight(s_k, TRUCK['w'])

    print("\n--- 🏁 ЕТАП 3: ВИЇЗД (2 камери) ---")
    send_cam(k['OUT'][0], TRUCK['f'], "EXIT Front")
    send_cam(k['OUT'][1], TRUCK['b'], "EXIT Back")
    send_cam(k['OUT'][1], TRUCK['b'], "EXIT Back")


    # Перевірка
    time.sleep(2)
    print("\n📊 ПЕРЕВІРКА РЕЗУЛЬТАТІВ:")
    h = get_headers(token)
    r = requests.get(f"{CORE_API_URL}/permits/?plate={TRUCK['f']}", headers=h).json()
    
    if r['data']:
        print("✅ Перепустку знайдено!")
        print(r['data'][0]["ID"])
        p = r['data'][0]
        print(f"   🚚 Фура: {p['plate_front']} / {p['plate_back']}")
        print(f"   ⚖️  Вага: {p['total_weight']} кг")
        print(f"   📸 Кількість подій (має бути 6): {len(p.get('plate_events', []))}")
        print(f"   🏁 Статус: {'✅ ЗАКРИТО' if p['is_closed'] else '❌ ВІДКРИТО'}")
    else:
        print("❌ Перепустку не знайдено!")

if __name__ == "__main__":
    main()