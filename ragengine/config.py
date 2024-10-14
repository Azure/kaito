# config.py
import os

EMBEDDING_TYPE = os.getenv("EMBEDDING_TYPE", "local")
EMBEDDING_URL = os.getenv("EMBEDDING_URL")

INFERENCE_URL = os.getenv("INFERENCE_URL", "http://52.190.41.209/chat")
INFERENCE_ACCESS_SECRET = os.getenv("AccessSecret", "default-inference-secret")
# RESPONSE_FIELD = os.getenv("RESPONSE_FIELD", "result")

MODEL_ID = os.getenv("MODEL_ID", "BAAI/bge-small-en-v1.5")
VECTOR_DB_TYPE = os.getenv("VECTOR_DB_TYPE", "faiss")
INDEX_SERVICE_NAME = os.getenv("INDEX_SERVICE_NAME", "default-index-service")
ACCESS_SECRET = os.getenv("ACCESS_SECRET", "default-access-secret")
PERSIST_DIR = "./storage"