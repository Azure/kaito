from typing import Dict, List

from models import Document
from vector_store.base import BaseVectorStore


class RAGOperations:
    def __init__(self, vector_store: BaseVectorStore):
        self.vector_store = vector_store

    def create(self, documents: List[Document]) -> List[str]:
        return self.vector_store.index_documents(documents)

    def read(self, query: str, top_k: int):
        return self.vector_store.query(query, top_k)

    def update(self, documents: List[Document]) -> Dict[str, List[str]]:
        updated_docs = []
        new_docs = []
        for doc in documents:
            if doc.doc_id and self.vector_store.document_exists(doc.doc_id):
                self.vector_store.update_document(doc)
                updated_docs.append(doc.doc_id)
            else:
                self.vector_store.add_document(doc)
                new_docs.extend(doc.doc_id)
        return {"updated": updated_docs, "inserted": new_docs}

    def delete(self, doc_id: str):
        return self.vector_store.delete_document(doc_id)

    def get(self, doc_id: str) -> Document:
        return self.vector_store.get_document(doc_id)

    def list_all(self) -> Dict[str, Document]:
        return self.vector_store.list_documents()

    def refresh(self, documents: List[Document]) -> List[bool]:
        return self.vector_store.refresh_documents(documents)
