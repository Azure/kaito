from abc import ABC, abstractmethod
from typing import Dict, List

from ragengine.models import Document
from llama_index.core import VectorStoreIndex
import hashlib


class BaseVectorStore(ABC):
    def generate_doc_id(text: str) -> str:
        """Generates a unique document ID based on the hash of the document text."""
        return hashlib.sha256(text.encode('utf-8')).hexdigest()

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
