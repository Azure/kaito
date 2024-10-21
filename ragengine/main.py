from typing import Dict, List

from llama_index.core.schema import TextNode

from vector_store_manager.manager import VectorStoreManager
from embedding.huggingface_local import LocalHuggingFaceEmbedding
from embedding.huggingface_remote import RemoteHuggingFaceEmbedding
from llama_index.core.storage.docstore.types import RefDocInfo
from fastapi import FastAPI, HTTPException
from models import (IndexRequest, ListDocumentsResponse,
                    QueryRequest, Document)
from vector_store.faiss_store import FaissVectorStoreHandler

from config import ACCESS_SECRET, EMBEDDING_TYPE, MODEL_ID

app = FastAPI()

# Initialize embedding model
if EMBEDDING_TYPE.lower() == "local":
    embedding_manager = LocalHuggingFaceEmbedding(MODEL_ID)
elif EMBEDDING_TYPE.lower() == "remote":
    embedding_manager = RemoteHuggingFaceEmbedding(MODEL_ID, ACCESS_SECRET)
else:
    raise ValueError("Invalid Embedding Type Specified (Must be Local or Remote)")

# Initialize vector store
# TODO: Dynamically set VectorStore from EnvVars (which ultimately comes from CRD StorageSpec)
vector_store_handler = FaissVectorStoreHandler(embedding_manager)

# Initialize RAG operations
rag_ops = VectorStoreManager(vector_store_handler)

@app.post("/index", response_model=List[Document])
async def index_documents(request: IndexRequest): # TODO: Research async/sync what to use (inference is calling)
    try:
        doc_ids = rag_ops.create(request.index_name, request.documents)
        documents = [
            Document(doc_id=doc_id, text=doc.text, metadata=doc.metadata)
            for doc_id, doc in zip(doc_ids, request.documents)
        ]
        return documents
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.post("/query", response_model=Dict[str, str])
async def query_index(request: QueryRequest):
    try:
        llm_params = request.llm_params or {} # Default to empty dict if no params provided
        response = rag_ops.read(request.index_name, request.query, request.top_k, llm_params)
        return {"response": str(response)}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/indexed-documents", response_model=ListDocumentsResponse)
async def list_all_indexed_documents():
    try:
        documents = rag_ops.list_all_indexed_documents()
        serialized_documents = {
            index_name: {
                doc_name: {
                    "text": doc_info.text, "hash": doc_info.hash
                } for doc_name, doc_info in vector_store_index.docstore.docs.items()
            }
            for index_name, vector_store_index in documents.items()
        }
        return ListDocumentsResponse(documents=serialized_documents)
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)