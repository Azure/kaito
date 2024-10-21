from typing import Dict, List

from ragengine.models import Document
from ragengine.vector_store.base import BaseVectorStore

from llama_index.core import VectorStoreIndex

class VectorStoreManager:
    def __init__(self, vector_store: BaseVectorStore):
        self.vector_store = vector_store

    def create(self, index_name: str, documents: List[Document]) -> List[str]:
        """Index new documents."""
        return self.vector_store.index_documents(index_name, documents)

    def read(self, index_name: str, query: str, top_k: int, llm_params: dict):
        """Query the indexed documents."""
        return self.vector_store.query(index_name, query, top_k, llm_params)

    def list_all_indexed_documents(self) -> Dict[str, VectorStoreIndex]:
        """List all documents."""
        return self.vector_store.list_all_indexed_documents()
