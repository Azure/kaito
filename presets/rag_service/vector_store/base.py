from abc import ABC, abstractmethod
from typing import Dict, List

from models import Document


class BaseVectorStore(ABC):
    @abstractmethod
    def index_documents(self, documents: List[Document]) -> List[str]:
        pass

    @abstractmethod
    def query(self, query: str, top_k: int):
        pass

    @abstractmethod
    def add_document(self, document: Document): 
        pass

    @abstractmethod
    def delete_document(self, doc_id: str):
        pass

    @abstractmethod
    def update_document(self, document: Document) -> str:
        pass

    @abstractmethod
    def get_document(self, doc_id: str) -> Document:
        pass

    @abstractmethod
    def list_documents(self) -> Dict[str, Document]:
        pass

    @abstractmethod
    def document_exists(self, doc_id: str) -> bool:
        pass

    @abstractmethod
    def refresh_documents(self, documents: List[Document]) -> List[bool]:
        pass

    @abstractmethod
    def list_documents(self) -> Dict[str, Document]:
        pass

    @abstractmethod
    def document_exists(self, doc_id: str) -> bool:
        pass