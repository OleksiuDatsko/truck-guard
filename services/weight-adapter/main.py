import time
import logging
from redis import Redis
from src.config import cfg
from src.clients.core_client import CoreClient
from src.logic.payload_parser import PayloadParser
from src.logic.processor import EventProcessor
from src.telemetry import init_telemetry

from src.utils.logging_utils import logger

def main():
    # Initialize OpenTelemetry
    init_telemetry("truckguard-weight-adapter")
    
    logger.info("Starting Weight Adapter Worker...")
    
    redis = Redis.from_url(f"redis://{cfg.REDIS_ADDR}", decode_responses=True) 
    core = CoreClient()
    parser = PayloadParser()
    processor = EventProcessor(core, parser)

    last_id = "0"
    while True:
        try:
            streams = redis.xread({cfg.STREAM_RAW: last_id}, count=1, block=5000)
            if not streams:
                continue

            for _, messages in streams:
                for msg_id, data in messages:
                    try:
                        processor.process(data["data"]) 
                    except Exception as e:
                        logger.error(f"Failed to process message {msg_id}: {e}")
                        redis.xadd(cfg.STREAM_DLQ, {"data": data["data"], "error": str(e)})
                    
                    last_id = msg_id

        except Exception as e:
            logger.critical(f"Redis connection error: {e}")
            time.sleep(5)

if __name__ == "__main__":
    main()
