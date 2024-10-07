from abc import ABC, abstractmethod
from typing import Dict, List

from models import Document


class BaseVectorStore(ABC):
    @abstractmethod
    def index_documents(self, documents: List[Document], index_name: str) -> List[str]:
        pass

    @abstractmethod
    def query(self, query: str, top_k: int, index_name: str, params: dict):
        pass

    @abstractmethod
    def add_document(self, document: Document, index_name: str): 
        pass

    """
    @abstractmethod
    def delete_document(self, doc_id: str, index_name: str):
        pass
        
    @abstractmethod
    def update_document(self, document: Document, index_name: str) -> str:
        pass

    @abstractmethod
    def refresh_documents(self, documents: List[Document], index_name: str) -> List[bool]:
        pass
    """

    @abstractmethod
    def get_document(self, doc_id: str, index_name: str) -> Document:
        pass

    @abstractmethod
    def list_documents(self, index_name: str) -> Dict[str, Document]:
        pass

    @abstractmethod
    def document_exists(self, doc_id: str, index_name: str) -> bool:
        pass

    @abstractmethod
    def list_documents(self, index_name: str) -> Dict[str, Document]:
        pass

    @abstractmethod
    def document_exists(self, doc_id: str, index_name: str) -> bool:
        pass