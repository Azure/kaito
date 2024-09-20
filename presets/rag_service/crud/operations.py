from typing import Dict, List

from models import Document
from vector_store.base import BaseVectorStore


class RAGOperations:
    def __init__(self, vector_store: BaseVectorStore):
        self.vector_store = vector_store

    def create(self, documents: List[Document]) -> List[str]:
        """Index new documents."""
        return self.vector_store.index_documents(documents)

    def read(self, query: str, top_k: int):
        """Query the indexed documents."""
        return self.vector_store.query(query, top_k)

    def update(self, documents: List[Document]) -> Dict[str, List[str]]:
        """Update existing documents, or insert new ones if they don’t exist."""
        updated_docs = []
        new_docs = []
        for doc in documents:
            if doc.doc_id and self.vector_store.document_exists(doc.doc_id):
                self.vector_store.update_document(doc)
                updated_docs.append(doc.doc_id)
            else:
                self.vector_store.add_document(doc)  # Only inserts new document, no reindex
                new_docs.append(doc.doc_id)
        return {"updated": updated_docs, "inserted": new_docs}

    def delete(self, doc_id: str):
        """Delete a document by ID."""
        return self.vector_store.delete_document(doc_id)

    def get(self, doc_id: str) -> Document:
        """Retrieve a document by ID."""
        return self.vector_store.get_document(doc_id)

    def list_all(self) -> Dict[str, Document]:
        """List all documents."""
        return self.vector_store.list_documents()

    def refresh(self, documents: List[Document]) -> List[bool]:
        """Dummy method for refresh, if needed."""
        return self.vector_store.refresh_documents(documents)
