import os

import faiss
from llama_index.core import Document as LlamaDocument
from llama_index.core import StorageContext, VectorStoreIndex
from llama_index.vector_stores.faiss import FaissVectorStore
from models import Document

from config import PERSIST_DIR

from .base import BaseVectorStore


class FaissVectorStoreManager(BaseVectorStore):
    def __init__(self, dimension: int, embed_model):
        self.dimension = dimension
        self.embed_model = embed_model
        self.faiss_index = faiss.IndexFlatL2(self.dimension)
        self.vector_store = FaissVectorStore(faiss_index=self.faiss_index)
        self.storage_context = StorageContext.from_defaults(vector_store=self.vector_store)
        
        if not os.path.exists(PERSIST_DIR):
            os.makedirs(PERSIST_DIR)

    def index_documents(self, documents: List[Document]):
        llama_docs = [LlamaDocument(text=doc.text, metadata=doc.metadata, id_=doc.doc_id) for doc in documents]
        index = VectorStoreIndex.from_documents(llama_docs, storage_context=self.storage_context, embed_model=self.embed_model)
        self.storage_context.persist(persist_dir=PERSIST_DIR)
        return index

    def query(self, query: str, top_k: int):
        index = self._load_index()
        query_engine = index.as_query_engine(top_k=top_k)
        return query_engine.query(query)
    
    def add_document(self, document: Document): 
        index = self._load_index()
        index.insert(document)

    def delete_document(self, doc_id: str):
        index = self._load_index()
        index.delete_ref_doc(doc_id, delete_from_docstore=True)
        self.storage_context.persist(persist_dir=PERSIST_DIR)

    def update_document(self, document: Document):
        index = self._load_index()
        llama_doc = LlamaDocument(text=document.text, metadata=document.metadata, id_=document.doc_id)
        index.update_ref_doc(llama_doc)
        self.storage_context.persist(persist_dir=PERSIST_DIR)

    def get_document(self, doc_id: str):
        index = self._load_index()
        doc = index.docstore.get_document(doc_id)
        if not doc:
            raise ValueError(f"Document with ID {doc_id} not found.")
        return doc

    def _load_index(self):
        vector_store = FaissVectorStore.from_persist_dir(PERSIST_DIR)
        storage_context = StorageContext.from_defaults(vector_store=vector_store, persist_dir=PERSIST_DIR)
        return VectorStoreIndex.from_storage(storage_context)
