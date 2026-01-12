import json
import xmltodict
from typing import Optional
from src.utils.logging_utils import logger

class PayloadParser:
    @staticmethod
    def extract_plate(payload: str, format_type: str, mapping: dict) -> Optional[str]:
        try:
            data = xmltodict.parse(payload) if format_type == "xml" else json.loads(payload)
            
            path = mapping.get("plate")
            if not path: return None

            keys = path.split('/')
            for k in keys:
                if isinstance(data, dict):
                    data = data.get(k)
            
            return str(data) if data else None
        except Exception as e:
            logger.warning(f"Failed to parse payload: {e}")
            return None