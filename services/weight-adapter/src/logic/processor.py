import json
import logging

logger = logging.getLogger("weight-adapter")
logging.basicConfig(level=logging.INFO)

class EventProcessor:
    def __init__(self, core_client, parser):
        self.core = core_client
        self.parser = parser
        self.config_cache = {}

    def _get_cached_config(self, source_id: str):
        if source_id in self.config_cache:
            return self.config_cache[source_id]
        config = self.core.get_scale_config(source_id)
        if config:
            self.config_cache[source_id] = config
        return config

    def process(self, raw_data_str: str):
        data = json.loads(raw_data_str)
        source_id = data.get("source_id")

        config = self._get_cached_config(source_id)
        if not config:
            logger.error(f"Config not found for scale {source_id}")
            return

        mapping = config.get("field_mapping", {})
        if isinstance(mapping, str):
            mapping = json.loads(mapping)

        weight = self.parser.extract_value(
            data.get("payload"), config.get("format"), mapping
        )

        if weight is not None:
            final_event = {
                "scale_id": source_id,
                "weight": weight,
                "timestamp": data.get("at"),
                "raw_payload": data.get("payload"),
            }
            try:
                self.core.send_weight_event(final_event)
                logger.info(f"Processed weight for {source_id}: {weight} kg")
            except Exception as e:
                logger.error(f"Failed to send weight event to Core: {e}")
                raise e
        else:
            logger.warning(f"Could not extract weight for {source_id}")
