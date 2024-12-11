# Copyright (c) Microsoft Corporation.
# Licensed under the MIT license.

# config.py

# Configuration variables are set via environment variables from the RAGEngine CR
# and exposed to the pod. For example, `LLM_INFERENCE_URL` is specified in the CR and
# passed to the pod via environment variables.

import os

# Embedding configuration
EMBEDDING_SOURCE_TYPE = os.getenv("EMBEDDING_SOURCE_TYPE", "local")  # Determines local or remote embedding source

# Local embedding model
LOCAL_EMBEDDING_MODEL_ID = os.getenv("LOCAL_EMBEDDING_MODEL_ID", "BAAI/bge-small-en-v1.5")

# Remote embedding model (if not local)
REMOTE_EMBEDDING_URL = os.getenv("REMOTE_EMBEDDING_URL", "http://localhost:5000/embedding")
REMOTE_EMBEDDING_ACCESS_SECRET = os.getenv("REMOTE_EMBEDDING_ACCESS_SECRET", "default-access-secret")

# LLM (Large Language Model) configuration
LLM_INFERENCE_URL = os.getenv("LLM_INFERENCE_URL", "http://localhost:5000/chat")
LLM_ACCESS_SECRET = os.getenv("LLM_ACCESS_SECRET", "default-access-secret")
# LLM_RESPONSE_FIELD = os.getenv("LLM_RESPONSE_FIELD", "result")  # Uncomment if needed in the future

# Vector database configuration
VECTOR_DB_IMPLEMENTATION = os.getenv("VECTOR_DB_IMPLEMENTATION", "faiss")
VECTOR_DB_PERSIST_DIR = os.getenv("VECTOR_DB_PERSIST_DIR", "storage")
