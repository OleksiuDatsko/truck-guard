import requests
import json
import time
import io
import os
import random
import string
from PIL import Image

# Конфігурація
BASE_URL = os.getenv("BASE_URL", "http://localhost")
AUTH_URL = f"{BASE_URL}/auth"
CORE_API_URL = f"{BASE_URL}/api"
INGEST_CAMERA_URL = f"{BASE_URL}/ingest/camera"
INGEST_WEIGHT_URL = f"{BASE_URL}/ingest/weight"

ADMIN_USER = os.getenv("ADMIN_USER", "admin")
ADMIN_PASS = os.getenv("ADMIN_DEFAULT_PASSWORD", "secret123")

TRUCK_COUNT = int(os.getenv("TRUCK_COUNT", "3"))

def get_headers(token):
    return {"Authorization": f"Bearer {token}"}

def cleanup(token):
    print("🧹 Повна очистка конфігурацій...")
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
    h = get_headers(token)
    print("🏗️ Створення інфраструктури...")
    
    g_in = requests.post(f"{CORE_API_URL}/configs/gates", headers=h, json={"name":"ENTRY", "is_entry":True}).json()['ID']
    g_sc = requests.post(f"{CORE_API_URL}/configs/gates", headers=h, json={"name":"SCALE"}).json()['ID']
    g_out = requests.post(f"{CORE_API_URL}/configs/gates", headers=h, json={"name":"EXIT", "is_exit":True}).json()['ID']
    
    env_keys = {
        "gate_ids": [g_in, g_sc, g_out],
        "cam_keys": {},
        "scale_key": ""
    }

    gate_configs = [
        ("IN", g_in),
        ("SC", g_sc),
        ("OUT", g_out)
    ]

    for prefix, g_id in gate_configs:
        key_f = requests.post(f"{CORE_API_URL}/configs/cameras", headers=h, 
                              json={"name": f"{prefix}_Front", "gate_id": g_id, "format": "json", "field_mapping": '{"plate":"plate"}'}).json()['api_key']
        key_b = requests.post(f"{CORE_API_URL}/configs/cameras", headers=h, 
                              json={"name": f"{prefix}_Back", "gate_id": g_id, "format": "json", "field_mapping": '{"plate":"plate"}'}).json()['api_key']
        env_keys["cam_keys"][prefix] = [key_f, key_b]

    s_key = requests.post(f"{CORE_API_URL}/configs/scales", headers=h, 
                          json={"name": "Main_Scale", "gate_id": g_sc, "format": "json", "field_mapping": '{"weight":"weight"}'}).json()['api_key']
    env_keys["scale_key"] = s_key

    # Setup Flow
    print("🌊 Налаштування Flow маршруту...")
    requests.post(f"{CORE_API_URL}/configs/flows", headers=h, json={
        "name": "Standard Flow",
        "steps": [
            {"gate_id": g_in, "sequence": 1},
            {"gate_id": g_sc, "sequence": 2},
            {"gate_id": g_out, "sequence": 3}
        ]
    })

    return env_keys

def send_cam(key, plate, cam_label=""):
    f = io.BytesIO()
    # Random color for variety
    Image.new('RGB', (100, 100), color=(random.randint(0,255), random.randint(0,255), random.randint(0,255))).save(f, 'jpeg')
    f.seek(0)
    requests.post(INGEST_CAMERA_URL, headers={'X-API-Key':key}, files={'image':('p.jpg',f)}, 
                  data={'device_id':'SIM','payload':json.dumps({"plate":plate})})
    print(f"   📸 {cam_label}: {plate}")

def send_weight(key, val, truck_plate):
    requests.post(INGEST_WEIGHT_URL, headers={'X-API-Key':key}, 
                  data={'device_id':'SCALE','payload':json.dumps({"weight":val})})
    print(f"   ⚖️  Вага для {truck_plate}: {val} kg")

def generate_plate():
    # Example format: AA1234BB
    letters = ''.join(random.choices(string.ascii_uppercase, k=2))
    nums = ''.join(random.choices(string.digits, k=4))
    letters2 = ''.join(random.choices(string.ascii_uppercase, k=2))
    return f"{letters}{nums}{letters2}"

def main():
    print(f"🚀 Запуск симуляції для {TRUCK_COUNT} вантажівок")
    
    # Auth
    token = requests.post(f"{AUTH_URL}/login", json={"username":ADMIN_USER, "password":ADMIN_PASS}).json().get("token")
    if not token: 
        print("❌ Не вдалося отримати токен")
        return

    cleanup(token)
    env = setup_env(token)
    keys = env['cam_keys']
    scale_key = env['scale_key']

    # Generate Trucks
    trucks = []
    base_start_time = time.time()
    
    print("\n⏱️  Генеруємо розклад руху...")
    for i in range(TRUCK_COUNT):
        p_front = generate_plate()
        p_back = generate_plate()
        weight = random.randint(15000, 40000)
        
        # Staggered start: Delay BETWEEN trucks
        start_delay = i * random.uniform(5.0, 10.0)
        
        # Steps delays (actions within truck)
        drive_time_1 = random.uniform(3.0, 6.0) # Entry -> Scale
        drive_time_2 = random.uniform(10.0, 20.0) # Scale -> Exit
        
        truck = {
            "id": i + 1,
            "plate_f": p_front,
            "plate_b": p_back,
            "weight": weight,
            "next_action_time": base_start_time + start_delay,
            "tasks": [
                # STEP 1: ENTRY
                ("CAM", keys['IN'][0], p_front, "ENTRY Front"),
                ("CAM", keys['IN'][1], p_back,  "ENTRY Back"),
                
                ("WAIT", drive_time_1),
                
                # STEP 2: SCALE
                ("CAM", keys['SC'][0], p_front, "SCALE Front"),
                ("CAM", keys['SC'][1], p_back,  "SCALE Back"),
                ("WEIGHT", scale_key, weight, ""),

                ("WAIT", drive_time_2),
                
                # STEP 3: EXIT
                ("CAM", keys['OUT'][0], p_front, "EXIT Front"),
                ("CAM", keys['OUT'][1], p_back,  "EXIT Back"), 
            ]
        }
        trucks.append(truck)
        print(f"🚛 Truck {truck['id']}: {truck['plate_f']} (Старт через: {start_delay:.1f}s)")

    print("\n🏁 Починаємо рух потік...")
    
    unfinished_trucks = [t for t in trucks if len(t['tasks']) > 0]
    
    while unfinished_trucks:
        now = time.time()
        # Find trucks ready to act
        ready = [t for t in unfinished_trucks if t['next_action_time'] <= now]
        
        if not ready:
            time.sleep(0.1)
            continue

        # Pick a random ready truck
        t = random.choice(ready)
        
        # Pop next task
        task = t['tasks'].pop(0)
        action_type = task[0]
        
        if action_type == "WAIT":
            wait_time = task[1]
            t['next_action_time'] = now + wait_time
            # print(f"   ⏳ Truck {t['id']} driving... ({wait_time:.1f}s)")
        
        elif action_type == "CAM":
            send_cam(task[1], task[2], f"[{t['id']}] {task[3]}")
            # Small natural delay between bursts of cams
            t['next_action_time'] = now + random.uniform(0.2, 0.5)
            
        elif action_type == "WEIGHT":
            send_weight(task[1], task[2], t['plate_f'])
            t['next_action_time'] = now + random.uniform(0.5, 1.0)
            
        if len(t['tasks']) == 0:
            unfinished_trucks.remove(t)
            print(f"🎉 Truck {t['id']} finished!")

    # Verify Results
    print("\n📊 ПЕРЕВІРКА РЕЗУЛЬТАТІВ:")
    time.sleep(2) # Wait for async processing
    h = get_headers(token)
    
    success_count = 0
    for t in trucks:
        print(f"\n🔎 Перевірка Truck {t['id']} ({t['plate_f']})...")
        r = requests.get(f"{CORE_API_URL}/permits/?plate={t['plate_f']}", headers=h).json()
        
        if r['data']:
            p = r['data'][0]
            # Verify weight
            w_diff = abs(p['total_weight'] - t['weight'])
            status_ok = p['is_closed']
            
            # Count events
            events_count = 0
            for ge in p.get('gate_events', []):
                 events_count += len(ge.get('plate_events', []))
                 events_count += len(ge.get('weight_events', []))
            print(f"    ✅  Перепустка знайдена. ID: {p['ID']}")
            print(f"    ⚖️  Вага: {p['total_weight']} (Очікувалось {t['weight']})")
            print(f"    ⏰  Час відкриття: {p['entry_time']}")
            print(f"    ⏰  Час закриття: {p['exit_time']}")
            print(f"    📸  Подій: {events_count} (Очікувалось ~7)")
            print(f"    🏁  Статус Closed: {status_ok}")
            
            if status_ok and w_diff < 1.0:
                success_count += 1
            else:
                print("   ⚠️  Щось не так з даними!")
        else:
            print("   ❌ Перепустку НЕ знайдено!")

    print(f"\n📈 Результат: {success_count}/{TRUCK_COUNT} успішних проїздів.")

if __name__ == "__main__":
    main()
