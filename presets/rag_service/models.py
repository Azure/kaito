from typing import Dict, List, Optional

from pydantic import BaseModel


class Document(BaseModel):
    text: str
    metadata: Optional[dict] = {}
    doc_id: Optional[str] = None

class IndexRequest(BaseModel):
    documents: List[Document]

class QueryRequest(BaseModel):
    query: str
    top_k: int = 10

class UpdateRequest(BaseModel):
    documents: List[Document]

class RefreshRequest(BaseModel):
    documents: List[Document]

class DocumentResponse(BaseModel):
    doc_id: str
    document: Document

class ListDocumentsResponse(BaseModel):
    documents: Dict[str, Document]