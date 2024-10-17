import os
from typing import Dict, List

import faiss
from llama_index.core import Document as LlamaDocument
from llama_index.core import (StorageContext, VectorStoreIndex)
from llama_index.core.storage.index_store import SimpleIndexStore
from llama_index.core.storage.docstore.types import RefDocInfo
from llama_index.vector_stores.faiss import FaissVectorStore

from ragengine.models import Document
from ragengine.inference.inference import Inference

from config import PERSIST_DIR

from .base import BaseVectorStore
from ragengine.embedding.base import BaseEmbeddingModel


class FaissVectorStoreHandler(BaseVectorStore):
    def __init__(self, embedding_manager: BaseEmbeddingModel):
        self.embedding_manager = embedding_manager
        self.embed_model =  self.embedding_manager.model
        self.dimension = self.embedding_manager.get_embedding_dimension()
        # TODO: Consider allowing user custom indexing method (would require configmap?) e.g.
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
        self.index_map = {} # Used to store the in-memory index via namespace (e.g. index_name -> VectorStoreIndex)
        self.index_store = SimpleIndexStore() # Use to store global index metadata
        self.llm = Inference()

    def index_documents(self, index_name: str, documents: List[Document]):
        """Recreates the entire FAISS index and vector store with new documents."""
        indexed_doc_ids = set()

        if index_name in self.index_map:
            print(f"Index {index_name} already exists. Appending documents to existing index.")
            for doc in documents:
                doc.doc_id = self.generate_doc_id(doc.text)
                if not self.document_exists(index_name, doc.doc_id):
                    self.add_document(index_name, doc)
                    indexed_doc_ids.add(doc.doc_id)
                else:
                    print(f"Document {doc.doc_id} already exists in index {index_name}. Skipping.")
            if indexed_doc_ids: 
                self._persist(index_name)
            return list(indexed_doc_ids)  # Return only the IDs of newly indexed documents

        faiss_index = faiss.IndexFlatL2(self.dimension) # Specifies FAISS indexing method (https://github.com/facebookresearch/faiss/wiki/Faiss-indexes)
        vector_store = FaissVectorStore(faiss_index=faiss_index) # Specifies in-memory data structure for storing and retrieving document embeddings
        storage_context = StorageContext.from_defaults(vector_store=vector_store) # Used to persist the vector store and its underlying data across sessions

        llama_docs = []
        for doc in documents:
            doc.doc_id = self.generate_doc_id(doc.text)
            # Check for duplicates before indexing
            if doc.doc_id not in indexed_doc_ids:
                llama_doc = LlamaDocument(id_=doc.doc_id, text=doc.text, metadata=doc.metadata)
                llama_docs.append(llama_doc)
                indexed_doc_ids.add(doc.doc_id)
            else:
                print(f"Document {doc.doc_id} already exists in index {index_name}. Skipping.")

        if llama_docs:
            # Creates the actual vector-based index using indexing method, vector store, storage method and embedding model specified above
            index = VectorStoreIndex.from_documents(
                llama_docs,
                storage_context=storage_context,
                embed_model=self.embed_model,
                # use_async=True # TODO: Indexing Process Performed Async
            )
            index.set_index_id(index_name) # https://github.com/run-llama/llama_index/blob/main/llama-index-core/llama_index/core/indices/base.py#L138-L154
            self.index_map[index_name] = index
            self.index_store.add_index_struct(index.index_struct)
            self._persist(index_name)
        # Return the document IDs that were successfully indexed
        return list(indexed_doc_ids)

    def add_document(self, index_name: str, document: Document):
        """Inserts a single document into the existing FAISS index."""
        if index_name not in self.index_map:
            raise ValueError(f"No such index: '{index_name}' exists.")
        llama_doc = LlamaDocument(text=document.text, metadata=document.metadata, id_=document.doc_id)
        self.index_map[index_name].insert(llama_doc)

    def query(self, index_name: str, query: str, top_k: int, llm_params: dict):
        """Queries the FAISS vector store."""
        if index_name not in self.index_map:
            raise ValueError(f"No such index: '{index_name}' exists.")
        self.llm.set_params(llm_params)

        query_engine = self.index_map[index_name].as_query_engine(llm=self.llm, similarity_top_k=top_k)
        return query_engine.query(query)

    def list_all_indexed_documents(self) -> Dict[str, VectorStoreIndex]:
        """Lists all documents in the vector store."""
        return self.index_map

    def document_exists(self, index_name: str, doc_id: str) -> bool:
        """Checks if a document exists in the vector store."""
        if index_name not in self.index_map:
            print(f"No such index: '{index_name}' exists in vector store.")
            return False
        return doc_id in self.index_map[index_name].ref_doc_info

    def _persist_all(self):
        self.index_store.persist(os.path.join(PERSIST_DIR, "store.json")) # Persist global index store
        for idx in self.index_store.index_structs():
            self._persist(idx.index_id)

    def _persist(self, index_name: str):
        """Saves the existing FAISS index to disk."""
        self.index_store.persist(os.path.join(PERSIST_DIR, "store.json")) # Persist global index store
        assert index_name in self.index_map, f"No such index: '{index_name}' exists."

        # Persist each index's storage context separately
        storage_context = self.index_map[index_name].storage_context
        storage_context.persist(
            persist_dir=os.path.join(PERSIST_DIR, index_name)
        )
