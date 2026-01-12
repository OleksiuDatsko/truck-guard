import os
from dataclasses import dataclass


@dataclass(frozen=True)
class Config:
    CORE_URL: str = os.getenv("CORE_URL", "http://gateway/api/v1")
    WORKER_API_KEY: str = os.getenv(
        "WORKER_API_KEY", "worker_internal_secret_2025_token"
    )
    ANPR_URL: str = os.getenv("ANPR_URL", "http://anpr:8000/recognize")
    REDIS_ADDR: str = os.getenv("REDIS_ADDR", "redis:6379")
    MINIO_ENDPOINT: str = os.getenv("MINIO_ENDPOINT", "minio:9000")
    MINIO_ACCESS_KEY: str = os.getenv("MINIO_ACCESS_KEY", "minioadmin")
    MINIO_SECRET_KEY: str = os.getenv("MINIO_SECRET_KEY", "minioadmin")
    MINIO_BUCKET: str = os.getenv("MINIO_BUCKET", "truck-images")

    STREAM_RAW: str = "camera:raw"
    STREAM_DLQ: str = "camera:dlq"
    CACHE_TTL: int = 60


cfg = Config()
