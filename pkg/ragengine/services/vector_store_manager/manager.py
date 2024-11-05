# Copyright (c) Microsoft Corporation.
# Licensed under the MIT license.

from typing import Dict, List

from services.models import Document
from services.vector_store.base import BaseVectorStore

class VectorStoreManager:
    def __init__(self, vector_store: BaseVectorStore):
        self.vector_store = vector_store

    def index(self, index_name: str, documents: List[Document]) -> List[str]:
        """Index new documents."""
        return self.vector_store.index_documents(index_name, documents)

    def query(self, index_name: str, query: str, top_k: int, llm_params: dict):
        """Query the indexed documents."""
        return self.vector_store.query(index_name, query, top_k, llm_params)

    def list_all_indexed_documents(self) -> Dict[str, Dict[str, Dict[str, str]]]:
        """List all documents."""
        return self.vector_store.list_all_indexed_documents()
