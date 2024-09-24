from typing import Dict, List

from crud.operations import RAGOperations
from embedding import get_embedding_model
from fastapi import FastAPI, HTTPException
from models import (DocumentResponse, IndexRequest, ListDocumentsResponse,
                    QueryRequest, RefreshRequest, UpdateRequest)
from vector_store.faiss_store import FaissVectorStoreManager

from config import ACCESS_SECRET, EMBEDDING_TYPE, MODEL_ID

app = FastAPI()

# Initialize embedding model
embed_model = get_embedding_model(EMBEDDING_TYPE, MODEL_ID, ACCESS_SECRET)

# Initialize vector store
vector_store = FaissVectorStoreManager(dimension=384, embed_model=embed_model)

# Initialize RAG operations
rag_ops = RAGOperations(vector_store)

@app.post("/index", response_model=List[str])
async def index_documents(request: IndexRequest):
    try:
        doc_ids = rag_ops.create(request.documents)
        return doc_ids
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.post("/query")
async def query_index(request: QueryRequest): # TODO: Research async/sync what to use (inference is calling)
    try:
        response = rag_ops.read(request.query, request.top_k)
        return {"response": str(response)}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.put("/update", response_model=Dict[str, List[str]])
async def update_documents(request: UpdateRequest):
    try:
        result = rag_ops.update(request.documents)
        return result
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.post("/refresh", response_model=List[bool])
async def refresh_documents(request: RefreshRequest):
    try:
        result = rag_ops.refresh(request.documents)
        return result
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.delete("/document/{doc_id}")
async def delete_document(doc_id: str):
    try:
        rag_ops.delete(doc_id)
        return {"message": "Document deleted successfully"}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/document/{doc_id}", response_model=DocumentResponse)
async def get_document(doc_id: str):
    try:
        document = rag_ops.get(doc_id)
        return DocumentResponse(doc_id=doc_id, document=document)
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/documents", response_model=ListDocumentsResponse)
async def list_documents():
    try:
        documents = rag_ops.list_all()
        return ListDocumentsResponse(documents=documents)
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)