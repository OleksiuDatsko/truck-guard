import os
from dataclasses import dataclass

@dataclass(frozen=True)
class Config:
    CORE_URL: str = os.getenv("CORE_URL", "http://gateway/api")
    WORKER_API_KEY: str = os.getenv("WORKER_API_KEY", "worker_internal_secret_2025_token")
    REDIS_ADDR: str = os.getenv("REDIS_ADDR", "redis:6379")

    STREAM_RAW: str = "weight:raw"
    STREAM_DLQ: str = "weight:dlq"
    CACHE_TTL: int = 60

cfg = Config()
