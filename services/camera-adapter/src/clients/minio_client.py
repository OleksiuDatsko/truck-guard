from minio import Minio
from src.config import cfg
from src.utils.logging_utils import logger

class MinioStorage:
    def __init__(self):
        self.client = Minio(
            cfg.MINIO_ENDPOINT,
            access_key=cfg.MINIO_ACCESS_KEY,
            secret_key=cfg.MINIO_SECRET_KEY,
            secure=False
        )
        self.bucket = cfg.MINIO_BUCKET

    def get_image(self, image_key: str) -> bytes:
        """Завантажує зображення з MinIO та повертає байти"""
        try:
            response = self.client.get_object(self.bucket, image_key)
            data = response.read()
            response.close()
            response.release_conn()
            return data
        except Exception as e:
            logger.error(f"MinIO error for {image_key}: {e}")
            raise