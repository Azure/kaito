from typing import Dict, List, Optional

from pydantic import BaseModel

class Document(BaseModel):
    text: str
    metadata: Optional[dict] = {}

class IndexRequest(BaseModel):
    index_name: str
    documents: List[Document]

class QueryRequest(BaseModel):
    index_name: str
    query: str
    top_k: int = 10
    llm_params: Optional[Dict] = None  # Accept a dictionary for parameters

class ListDocumentsResponse(BaseModel):
    documents:Dict[str, Dict[str, Dict[str, str]]]
