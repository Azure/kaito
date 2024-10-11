import os
from tempfile import TemporaryDirectory
from unittest.mock import patch

import pytest
from vector_store.faiss_store import FaissVectorStoreHandler
from models import Document
from embedding.huggingface_local import LocalHuggingFaceEmbedding
from config import MODEL_ID, INFERENCE_URL, INFERENCE_ACCESS_SECRET

from main import app, rag_ops
from fastapi.testclient import TestClient
from unittest.mock import MagicMock

AUTO_GEN_DOC_ID_LEN = 36

client = TestClient(app)

def test_index_documents_success():
    request_data = {
        "index_name": "test_index",
        "documents": [
            {"text": "This is a test document"},
            {"text": "Another test document"}
        ]
    }

    response = client.post("/index", json=request_data)
    assert response.status_code == 200
    doc1, doc2 = response.json()
    assert (doc1["text"] == "This is a test document")
    assert len(doc1["doc_id"]) == AUTO_GEN_DOC_ID_LEN
    assert not doc1["metadata"]

    assert (doc2["text"] == "Another test document")
    assert len(doc2["doc_id"]) == AUTO_GEN_DOC_ID_LEN
    assert not doc2["metadata"]

@patch('requests.post')
def test_query_index_success(mock_post):
    # Define Mock Response for Custom Inference API
    mock_response = {
        "result": "This is the completion from the API"
    }
    mock_post.return_value.json.return_value = mock_response
     # Index
    request_data = {
        "index_name": "test_index",
        "documents": [
            {"text": "This is a test document"},
            {"text": "Another test document"}
        ]
    }

    response = client.post("/index", json=request_data)
    assert response.status_code == 200

    # Query
    request_data = {
        "index_name": "test_index",
        "query": "test query",
        "top_k": 1,
        "llm_params": {"temperature": 0.7}
    }

    response = client.post("/query", json=request_data)
    assert response.status_code == 200
    assert response.json() == {"response": "This is the completion from the API"}
    assert mock_post.call_count == 1

def test_query_index_failure():
    # Prepare request data for querying.
    request_data = {
        "index_name": "non_existent_index",  # Use an index name that doesn't exist
        "query": "test query",
        "top_k": 1,
        "llm_params": {"temperature": 0.7}
    }

    response = client.post("/query", json=request_data)
    assert response.status_code == 500
    assert response.json()["detail"] == "No such index: 'non_existent_index' exists."


def test_get_document_success():
    request_data = {
        "index_name": "test_index",
        "documents": [
            # {"doc_id": "doc1", "text": "This is a test document"},
            {"doc_id": "doc1", "text": "This is a test document"},
            {"text": "Another test document"}
        ]
    }

    index_response = client.post("/index", json=request_data)
    assert index_response.status_code == 200

    # Call the GET document endpoint.
    get_response = client.get("/document/test_index/doc1")
    assert get_response.status_code == 200

    response_json = get_response.json()

    assert response_json.keys() == {"node_ids", 'metadata'}
    assert response_json['metadata'] == {}

    assert isinstance(response_json["node_ids"], list) and len(response_json["node_ids"]) == 1


def test_get_document_failure():
    # Call the GET document endpoint.
    response = client.get("/document/test_index/doc1")
    assert response.status_code == 404

def test_list_all_indexed_documents_success():
    response = client.get("/indexed-documents")
    assert response.status_code == 200
    assert response.json() == {'documents': {}}

    request_data = {
        "index_name": "test_index",
        "documents": [
            {"text": "This is a test document"},
            {"text": "Another test document"}
        ]
    }

    response = client.post("/index", json=request_data)
    assert response.status_code == 200

    response = client.get("/indexed-documents")
    assert response.status_code == 200
    assert "test_index" in response.json()["documents"]
    response_idx = response.json()["documents"]["test_index"]
    assert len(response_idx) == 2 # Two Documents Indexed
    assert ({item["text"] for item in response_idx.values()}
            == {item["text"] for item in request_data["documents"]})

