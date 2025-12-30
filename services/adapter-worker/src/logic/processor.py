import json
from src.config import cfg
from src.utils.logging_utils import logger


class EventProcessor:
    def __init__(self, core_client, parser, minio_client, anpr_client):
        self.core = core_client
        self.parser = parser
        self.minio = minio_client
        self.anpr = anpr_client
        self.config_cache = {}

    def _get_cached_config(self, camera_id: str):
        if camera_id in self.config_cache:
            return self.config_cache[camera_id]

        config = self.core.get_camera_config(camera_id)
        if config:
            self.config_cache[camera_id] = config
        return config

    def process(self, raw_data_str: str):
        data = json.loads(raw_data_str)
        camera_id = data.get("camera_id")

        config = self._get_cached_config(camera_id)
        if not config:
            raise ValueError(f"Config not found for camera {camera_id}")

        mapping = config.get("field_mapping", {})
        if isinstance(mapping, str):
            mapping = json.loads(mapping)

        plate = self.parser.extract_plate(
            data.get("payload"), config.get("format"), mapping
        )

        if not plate or config.get("run_anpr"):
            logger.info(f"Running AI recognition for {data.get('image_key')}")
            #TODO Тут логіка виклику ANPR клієнта
            pass

        if plate:
            final_event = {
                "camera_id": camera_id,
                "camera_name": config.get("name", camera_id),
                "plate": plate.upper().replace(" ", ""),
                "image_key": data.get("image_key"),
                "timestamp": data.get("at"),
                "raw_payload": data.get("payload"),
            }
            self.core.send_event(final_event)
            logger.info(f"Successfully processed plate: {plate}")
