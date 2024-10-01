import os
from typing import Dict, List

import faiss
from llama_index.core import Document as LlamaDocument
from llama_index.core import (StorageContext, VectorStoreIndex,
                              load_graph_from_storage, load_index_from_storage,
                              load_indices_from_storage)
from llama_index.core.storage.index_store import SimpleIndexStore
from llama_index.vector_stores.faiss import FaissVectorStore
from llama_index.core.data_structs.data_structs import IndexStruct
from models import Document

from config import PERSIST_DIR

from .base import BaseVectorStore


class FaissVectorStoreManager(BaseVectorStore):
    def __init__(self, embedding_manager):
        self.embedding_manager = embedding_manager
        self.embed_model =  self.embedding_manager.model
        self.dimension = self.embedding_manager.get_embedding_dimension()
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
        self.index_map = {} # Used to store the in-memory index via namespace (e.g. namespace -> index)
        self.index_store = SimpleIndexStore() # Use to store global index metadata

    def index_documents(self, documents: List[Document], index_name: str):
        """Recreates the entire FAISS index and vector store with new documents."""
        if index_name in self.index_map:
            del self.index_map[index_name]
            self.index_store.delete_index_struct(self.index_map[index_name])
            print(f"Index {index_name} already exists. Overwriting.")

        faiss_index = faiss.IndexFlatL2(self.dimension) # Specifies FAISS indexing method (https://github.com/facebookresearch/faiss/wiki/Faiss-indexes)
        vector_store = FaissVectorStore(faiss_index=faiss_index) # Specifies in-memory data structure for storing and retrieving document embeddings
        storage_context = StorageContext.from_defaults(vector_store=vector_store) # Used to persist the vector store and its underlying data across sessions

        llama_docs = [LlamaDocument(text=doc.text, metadata=doc.metadata, id_=doc.doc_id) for doc in documents]
        # Creates the actual vector-based index using indexing method, vector store, storage method and embedding model specified above
        index = VectorStoreIndex.from_documents(
            llama_docs,
            storage_context=storage_context,
            embed_model=self.embed_model,
            use_async=True # Indexing Process Performed Async
        )
        index.set_index_id(index_name) # https://github.com/run-llama/llama_index/blob/main/llama-index-core/llama_index/core/indices/base.py#L138-L154
        self.index_map[index_name] = index
        self.index_store.add_index_struct(index.index_struct)
        self._persist(index_name) # TODO: Consider just persisting the index as opposed to shared index_store
        # Return the document IDs that were indexed
        return [doc.doc_id for doc in documents]

    def add_document(self, document: Document, index_name: str):
        """Inserts a single document into the existing FAISS index."""
        if index_name not in self.index_map:
            raise ValueError(f"No such index: '{index_name}' exists.")
        llama_doc = LlamaDocument(text=document.text, metadata=document.metadata, id_=document.doc_id)
        self.index_map[index_name].insert(llama_doc)
        self._persist(index_name)

    def query(self, query: str, top_k: int, index_name: str):
        """Queries the FAISS vector store."""
        if index_name not in self.index_map:
            raise ValueError(f"No such index: '{index_name}' exists.")
        query_engine = self.index_map[index_name].as_query_engine(top_k=top_k)
        return query_engine.query(query)

    def delete_document(self, doc_id: str, index_name: str):
        """Deletes a document from the FAISS vector store."""
        if index_name not in self.index_map:
            raise ValueError(f"No such index: '{index_name}' exists.")
        self.index_map[index_name].delete_ref_doc(doc_id, delete_from_docstore=True)
        self._persist(index_name)

    def update_document(self, document: Document, index_name: str):
        """Updates an existing document in the FAISS vector store."""
        if index_name not in self.index_map:
            raise ValueError(f"No such index: '{index_name}' exists.")
        llama_doc = LlamaDocument(text=document.text, metadata=document.metadata, id_=document.doc_id)
        self.index_map[index_name].update_ref_doc(llama_doc)
        self._persist(index_name)

    def get_document(self, doc_id: str, index_name: str):
        """Retrieves a document's RefDocInfo by its ID."""
        if index_name not in self.index_map:
            raise ValueError(f"No such index: '{index_name}' exists.")

        # Try to retrieve the RefDocInfo associated with the doc_id
        ref_doc_info = self.index_map[index_name].ref_doc_info.get(doc_id)

        if ref_doc_info is None:
            print(f"Document with ID {doc_id} not found in index '{index_name}'.")
            return None

        return ref_doc_info

    def get_nodes_by_ref_doc_id(self, doc_id: str, index_name: str):
        """Retrieve nodes associated with a given document's ref ID."""
        if index_name not in self.index_map:
            raise ValueError(f"No such index: '{index_name}' exists.")

        ref_doc_info = self.get_document(doc_id, index_name)
        if ref_doc_info is None:
            return None

        return ref_doc_info.node_ids

    def refresh_documents(self, documents: List[Document], index_name: str) -> List[bool]:
        """Updates existing documents and inserts new documents in the vector store."""
        if index_name not in self.index_map:
            raise ValueError(f"No such index: '{index_name}' exists.")
        llama_docs = [LlamaDocument(text=doc.text, metadata=doc.metadata, id_=doc.doc_id) for doc in documents]
        refresh_results = self.index_map[index_name].refresh_ref_docs(llama_docs)
        self._persist(index_name)
        # Returns a list of booleans indicating whether each document was successfully refreshed.
        return refresh_results

    def list_documents(self, index_name: str) -> Dict[str, Document]:
        """Lists all documents in the vector store."""
        if index_name not in self.index_map:
            raise ValueError(f"No such index: '{index_name}' exists.")
        pass
        # return {doc_id: Document(text=doc.text, metadata=doc.metadata, doc_id=doc_id)
        #         for doc_id, doc in self.index_map[index_name].docstore.docs.items()}

    def document_exists(self, doc_id: str, index_name: str) -> bool:
        """Checks if a document exists in the vector store."""
        if index_name not in self.index_map:
            raise ValueError(f"No such index: '{index_name}' exists.")
        return doc_id in self.index_map[index_name].ref_doc_info

    def _load_index(self, index_name: str):
        """Loads the existing FAISS index from disk."""
        persist_dir = os.path.join(PERSIST_DIR, index_name)
        if not os.path.exists(persist_dir):
            raise ValueError(f"No persisted index found for '{index_name}'")
        vector_store = FaissVectorStore.from_persist_dir(persist_dir)
        storage_context = StorageContext.from_defaults(
            vector_store=vector_store, persist_dir=persist_dir
        )
        self.index_map[index_name] = load_index_from_storage(storage_context=storage_context)
        return self.index_map[index_name]

    def _persist(self, index_name: str):
        """Saves the existing FAISS index to disk."""
        self.index_store.persist(os.path.join(PERSIST_DIR, "store.json")) # Persist global index store
        assert index_name in self.index_map, f"No such index: '{index_name}' exists."

        # Persist each index's storage context separately
        storage_context = self.index_map[index_name].storage_context
        storage_context.persist(persist_dir=os.path.join(PERSIST_DIR, index_name))
