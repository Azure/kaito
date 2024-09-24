import os
from typing import Dict, List

import chromadb
from llama_index.core import Document as LlamaDocument
from llama_index.core import (StorageContext, VectorStoreIndex,
                              load_index_from_storage)
from llama_index.vector_stores.chroma import ChromaVectorStore
from models import Document

from config import PERSIST_DIR

from .base import BaseVectorStore


class ChromaDBVectorStoreManager(BaseVectorStore):
    def __init__(self, embed_model):
        self.embed_model = embed_model
        # Initialize ChromaDB client and collection
        self.chroma_client = chromadb.EphemeralClient()
        self.collection_name = "quickstart"
        self.chroma_collection = self.chroma_client.create_collection(self.collection_name)
        self.vector_store = ChromaVectorStore(chroma_collection=self.chroma_collection)
        self.storage_context = StorageContext.from_defaults(vector_store=self.vector_store)
        self.index = None  # Use to store the in-memory index # TODO: Multiple indexes via name (e.g. namespace)

        if not os.path.exists(PERSIST_DIR):
            os.makedirs(PERSIST_DIR)

    def index_documents(self, documents: List[Document]):
        """Recreates the entire ChromaDB index and vector store with new documents."""
        llama_docs = [LlamaDocument(text=doc.text, metadata=doc.metadata, id_=doc.doc_id) for doc in documents]
        self.index = VectorStoreIndex.from_documents(llama_docs, storage_context=self.storage_context, embed_model=self.embed_model)
        self._persist()
        # Return the document IDs that were indexed
        return [doc.doc_id for doc in documents]

    def add_document(self, document: Document):
        """Inserts a single document into the existing ChromaDB index."""
        if self.index is None:
            self.index = self._load_index()  # Load if not already in memory
        llama_doc = LlamaDocument(text=document.text, metadata=document.metadata, id_=document.doc_id)
        self.index.insert(llama_doc)
        self.storage_context.persist(persist_dir=PERSIST_DIR)

    def query(self, query: str, top_k: int):
        """Queries the ChromaDB vector store."""
        if self.index is None:
            self.index = self._load_index()  # Load if not already in memory
        query_engine = self.index.as_query_engine(top_k=top_k)
        return query_engine.query(query)

    def delete_document(self, doc_id: str):
        """Deletes a document from the ChromaDB vector store."""
        if self.index is None:
            self.index = self._load_index()  # Load if not already in memory
        self.index.delete_ref_doc(doc_id, delete_from_docstore=True)
        self.storage_context.persist(persist_dir=PERSIST_DIR)

    def update_document(self, document: Document):
        """Updates an existing document in the ChromaDB vector store."""
        if self.index is None:
            self.index = self._load_index()  # Load if not already in memory
        llama_doc = LlamaDocument(text=document.text, metadata=document.metadata, id_=document.doc_id)
        self.index.update_ref_doc(llama_doc)
        self.storage_context.persist(persist_dir=PERSIST_DIR)

    def get_document(self, doc_id: str):
        """Retrieves a document by its ID from ChromaDB."""
        if self.index is None:
            self.index = self._load_index()  # Load if not already in memory
        doc = self.index.docstore.get_document(doc_id)
        if not doc:
            raise ValueError(f"Document with ID {doc_id} not found.")
        return doc

    def refresh_documents(self, documents: List[Document]) -> List[bool]:
        """Updates existing documents and inserts new documents in the vector store."""
        if self.index is None:
            self.index = self._load_index()  # Load if not already in memory
        llama_docs = [LlamaDocument(text=doc.text, metadata=doc.metadata, id_=doc.doc_id) for doc in documents]
        refresh_results = self.index.refresh_ref_docs(llama_docs)
        self._persist()
        # Returns a list of booleans indicating whether each document was successfully refreshed.
        return refresh_results

    def list_documents(self) -> Dict[str, Document]:
        """Lists all documents in the ChromaDB vector store."""
        if self.index is None:
            self.index = self._load_index()  # Load if not already in memory
        return {doc_id: Document(text=doc.text, metadata=doc.metadata, doc_id=doc_id) 
                for doc_id, doc in self.index.docstore.docs.items()}

    def document_exists(self, doc_id: str) -> bool:
        """Checks if a document exists in the ChromaDB vector store."""
        if self.index is None:
            self.index = self._load_index()  # Load if not already in memory
        return doc_id in self.index.docstore.docs

    def _load_index(self):
        """Loads the existing ChromaDB index from disk."""
        vector_store = ChromaVectorStore(chroma_collection=self.chroma_collection)
        storage_context = StorageContext.from_defaults(
            vector_store=vector_store, persist_dir=PERSIST_DIR
        )
        return load_index_from_storage(storage_context=storage_context)

    def _persist(self):
        """Saves the existing ChromaDB index to disk."""
        self.storage_context.persist(persist_dir=PERSIST_DIR)
