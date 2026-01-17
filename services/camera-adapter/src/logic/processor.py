import json
from src.utils.logging_utils import logger


class EventProcessor:
    def __init__(self, core_client, parser, minio_client, anpr_client):
        self.core = core_client
        self.parser = parser
        self.minio = minio_client
        self.anpr = anpr_client
        self.config_cache = {}

    def _get_cached_config(self, source_id: str):
        if source_id in self.config_cache:
            return self.config_cache[source_id]
        config = self.core.get_camera_config(source_id)
        if config:
            self.config_cache[source_id] = config
        return config

    def process(self, raw_data_str: str):
        data = json.loads(raw_data_str)
        source_id = data.get("source_id")
        image_key = data.get("image_key")

        config = self._get_cached_config(source_id)
        if not config:
            raise ValueError(f"Config not found for source {source_id}")

        mapping = config.get("field_mapping", {})
        if isinstance(mapping, str):
            mapping = json.loads(mapping)

        plate = self.parser.extract_plate(
            data.get("payload"), config.get("format"), mapping
        )

        suggestions = []
        if not plate or config.get("run_anpr"):
            try:
                img_bytes = self.minio.get_image(image_key)
                suggestions = self.anpr.recognize(img_bytes)
                if suggestions and not plate:
                    plate = suggestions[0]["plate"]
            except Exception as e:
                logger.error(f"AI Recognition failed: {e}")

        if plate:
            final_event = {
                "camera_source_id": source_id,
                "camera_source_name": config.get("name", source_id),
                "camera_id": f"{config.get("ID")}",
                "plate": plate.upper().replace(" ", ""),
                "suggestions": json.dumps(suggestions), 
                "image_key": image_key,
                "timestamp": data.get("at"),
                "raw_payload": data.get("payload"),
            }
            self.core.send_event(final_event)
            logger.info(f"Successfully processed plate: {final_event}")
