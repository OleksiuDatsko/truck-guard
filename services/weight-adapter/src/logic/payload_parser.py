import json
import xmltodict
from typing import Optional

class PayloadParser:
    @staticmethod
    def extract_value(payload: str, format_type: str, mapping: dict) -> Optional[float]:
        try:
            data = xmltodict.parse(payload) if format_type == "xml" else json.loads(payload)
            
            path = mapping.get("weight")
            if not path: return None

            keys = path.split('/')
            for k in keys:
                if isinstance(data, dict):
                    data = data.get(k)
            
            return str(data) if data is not None else None
        except Exception as e:
            logger.warning(f"Failed to parse payload: {e}")
            return None
