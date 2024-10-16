from abc import ABC, abstractmethod
from typing import Dict, List

from models import Document
from llama_index.core import VectorStoreIndex


class BaseVectorStore(ABC):
    @abstractmethod
    def index_documents(self, index_name: str, documents: List[Document]) -> List[str]:
        pass

    @abstractmethod
    def query(self, index_name: str, query: str, top_k: int, params: dict):
        pass

    @abstractmethod
    def add_document(self, index_name: str, document: Document):
        pass

    @abstractmethod
    def list_all_indexed_documents(self) -> Dict[str, VectorStoreIndex]:
        pass

    @abstractmethod
    def document_exists(self, index_name: str, doc_id: str) -> bool:
        pass
