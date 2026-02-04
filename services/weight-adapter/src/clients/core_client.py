import requests
from src.config import cfg
from src.utils.logging_utils import logger
from tenacity import retry, stop_after_attempt, wait_exponential

class CoreClient:
    def __init__(self):
        self.headers = {
            "X-API-Key": cfg.WORKER_API_KEY, 
            "Content-Type": "application/json"
        }
    def get_scale_config(self, source_id: str) -> dict | None:
        url = f"{cfg.CORE_URL}/scales/by-id/{source_id}"
        try:
            resp = requests.get(url, headers=self.headers, timeout=5)
            if resp.status_code == 200:
                return resp.json()
        except Exception as e:
            logger.error("Error fetching scale config", extra={"source_id": source_id, "error": str(e)})
        return None

    @retry(stop=stop_after_attempt(5), wait=wait_exponential(multiplier=1, min=2, max=10))
    def send_weight_event(self, event_data: dict):
        url = f"{cfg.CORE_URL}/events/weight" 
        resp = requests.post(url, json=event_data, headers=self.headers, timeout=5) 
        resp.raise_for_status()
