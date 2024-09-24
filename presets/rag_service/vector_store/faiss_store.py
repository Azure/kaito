import os
from typing import Dict, List

import faiss
from llama_index.core import Document as LlamaDocument
from llama_index.core import (StorageContext, VectorStoreIndex,
                              load_graph_from_storage, load_index_from_storage,
                              load_indices_from_storage)
from llama_index.vector_stores.faiss import FaissVectorStore
from models import Document

from config import PERSIST_DIR

from .base import BaseVectorStore


class FaissVectorStoreManager(BaseVectorStore):
    def __init__(self, embed_model):
        self.embed_model = embed_model
        self.dimension = self.embed_model.get_embedding_dimension()
        # TODO: Consider allowing user custom indexing method e.g.
        """
        # Choose the FAISS index type based on the provided index_method
        if index_method == 'FlatL2':
            faiss_index = faiss.IndexFlatL2(self.dimension)  # L2 (Euclidean distance) index
        elif index_method == 'FlatIP':
            faiss_index = faiss.IndexFlatIP(self.dimension)  # Inner product (cosine similarity) index
        elif index_method == 'IVFFlat':
            quantizer = faiss.IndexFlatL2(self.dimension)  # Quantizer for IVF
            faiss_index = faiss.IndexIVFFlat(quantizer, self.dimension, 100)  # IVF with flat quantization
        elif index_method == 'HNSW':
            faiss_index = faiss.IndexHNSWFlat(self.dimension, 32)  # HNSW index with 32 neighbors
        else:
            raise ValueError(f"Unknown index method: {index_method}")
        """
        # TODO: We need to test if sharing storage_context is viable/correct or if we should make a new one for each index
        self.faiss_index = faiss.IndexFlatL2(self.dimension) # Specifies FAISS indexing method (https://github.com/facebookresearch/faiss/wiki/Faiss-indexes)
        self.vector_store = FaissVectorStore(faiss_index=self.faiss_index) # Specifies in-memory data structure for storing and retrieving document embeddings
        self.storage_context = StorageContext.from_defaults(vector_store=self.vector_store) # Used to persist the vector store and its underlying data across sessions
        self.indices = {} # Use to store the in-memory index via namespace (e.g. namespace -> index)

        if not os.path.exists(PERSIST_DIR):
            os.makedirs(PERSIST_DIR)

    def index_documents(self, documents: List[Document], index_name: str):
        """Recreates the entire FAISS index and vector store with new documents."""
        if index_name in self.indices:
            print(f"Index {index_name} already exists. Overwriting.")
        llama_docs = [LlamaDocument(text=doc.text, metadata=doc.metadata, id_=doc.doc_id) for doc in documents]
        # Creates the actual vector-based index using indexing method, vector store, storage method and embedding model specified above
        self.indices[index_name] = VectorStoreIndex.from_documents(llama_docs, storage_context=self.storage_context, embed_model=self.embed_model) 
        self._persist(index_name)
        # Return the document IDs that were indexed
        return [doc.doc_id for doc in documents]

    def add_document(self, document: Document, index_name: str):
        """Inserts a single document into the existing FAISS index."""
        assert index_name in self.indices, f"No such index: '{index_name}' exists."
        llama_doc = LlamaDocument(text=document.text, metadata=document.metadata, id_=document.doc_id)
        self.indices[index_name].insert(llama_doc)
        self.indices[index_name].storage_context.persist(persist_dir=PERSIST_DIR)

    def query(self, query: str, top_k: int, index_name: str):
        """Queries the FAISS vector store."""
        assert index_name in self.indices, f"No such index: '{index_name}' exists."
        query_engine = self.indices[index_name].as_query_engine(top_k=top_k)
        return query_engine.query(query)

    def delete_document(self, doc_id: str, index_name: str):
        """Deletes a document from the FAISS vector store."""
        assert index_name in self.indices, f"No such index: '{index_name}' exists."
        self.indices[index_name].delete_ref_doc(doc_id, delete_from_docstore=True)
        self.indices[index_name].storage_context.persist(persist_dir=PERSIST_DIR)

    def update_document(self, document: Document, index_name: str):
        """Updates an existing document in the FAISS vector store."""
        assert index_name in self.indices, f"No such index: '{index_name}' exists."
        llama_doc = LlamaDocument(text=document.text, metadata=document.metadata, id_=document.doc_id)
        self.indices[index_name].update_ref_doc(llama_doc)
        self.indices[index_name].storage_context.persist(persist_dir=PERSIST_DIR)

    def get_document(self, doc_id: str, index_name: str):
        """Retrieves a document by its ID."""
        assert index_name in self.indices, f"No such index: '{index_name}' exists."
        doc = self.indices[index_name].docstore.get_document(doc_id)
        if not doc:
            raise ValueError(f"Document with ID {doc_id} not found.")
        return doc

    def refresh_documents(self, documents: List[Document], index_name: str) -> List[bool]:
        """Updates existing documents and inserts new documents in the vector store."""
        assert index_name in self.indices, f"No such index: '{index_name}' exists."
        llama_docs = [LlamaDocument(text=doc.text, metadata=doc.metadata, id_=doc.doc_id) for doc in documents]
        refresh_results = self.indices[index_name].refresh_ref_docs(llama_docs)
        self._persist(index_name)
        # Returns a list of booleans indicating whether each document was successfully refreshed.
        return refresh_results

    def list_documents(self, index_name: str) -> Dict[str, Document]:
        """Lists all documents in the vector store."""
        assert index_name in self.indices, f"No such index: '{index_name}' exists."
        return {doc_id: Document(text=doc.text, metadata=doc.metadata, doc_id=doc_id) 
                for doc_id, doc in self.indices[index_name].docstore.docs.items()}

    def document_exists(self, doc_id: str, index_name: str) -> bool:
        """Checks if a document exists in the vector store."""
        assert index_name in self.indices, f"No such index: '{index_name}' exists."
        return doc_id in self.indices[index_name].docstore.docs

    def _load_index(self, index_name: str):
        """Loads the existing FAISS index from disk."""
        persist_dir = os.path.join(PERSIST_DIR, index_name)
        if not os.path.exists(persist_dir):
            raise ValueError(f"No persisted index found for '{index_name}'")
        vector_store = FaissVectorStore.from_persist_dir(persist_dir)
        storage_context = StorageContext.from_defaults(
            vector_store=vector_store, persist_dir=persist_dir
        )
        self.indices[index_name] = load_index_from_storage(storage_context=storage_context)
        return self.indices[index_name]

    def _persist(self, index_name: str):
        """Saves the existing FAISS index to disk."""
        assert index_name in self.indices, f"No such index: '{index_name}' exists."
        storage_context = self.indices[index_name].storage_context
        storage_context.persist(persist_dir=os.path.join(PERSIST_DIR, index_name))
