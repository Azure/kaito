import os
from tempfile import TemporaryDirectory
from unittest.mock import MagicMock

import pytest
from vector_store.faiss_store import FaissVectorStoreManager
from models import Document
from embedding.huggingface_local import LocalHuggingFaceEmbedding
from config import MODEL_ID

@pytest.fixture(scope='session')
def init_embed_manager():
    return LocalHuggingFaceEmbedding(MODEL_ID)

@pytest.fixture
def vector_store_manager(init_embed_manager):
    with TemporaryDirectory() as temp_dir:
        # Mock the persistence directory
        os.environ['PERSIST_DIR'] = temp_dir
        yield FaissVectorStoreManager(init_embed_manager)


def test_index_documents(vector_store_manager):
    documents = [
        Document(doc_id="1", text="First document", metadata={"type": "text"}),
        Document(doc_id="2", text="Second document", metadata={"type": "text"})
    ]
    
    doc_ids = vector_store_manager.index_documents(documents, index_name="test_index")
    
    assert len(doc_ids) == 2
    assert doc_ids == ["1", "2"]


def test_query_documents(vector_store_manager):
    # Add documents to index
    documents = [
        Document(doc_id="1", text="First document", metadata={"type": "text"}),
        Document(doc_id="2", text="Second document", metadata={"type": "text"})
    ]
    vector_store_manager.index_documents(documents, index_name="test_index")

    # Mock query and results
    query_result = vector_store_manager.query("First", top_k=1, index_name="test_index")
    
    assert query_result is not None


def test_add_and_delete_document(vector_store_manager):
    document = Document(doc_id="3", text="Third document", metadata={"type": "text"})
    vector_store_manager.index_documents([document], index_name="test_index")

    # Add a document to the existing index
    new_document = Document(doc_id="4", text="Fourth document", metadata={"type": "text"})
    vector_store_manager.add_document(new_document, index_name="test_index")

    # Assert that the document exists
    assert vector_store_manager.document_exists("4", "test_index")

    # Delete the document
    vector_store_manager.delete_document("4", "test_index")

    # Assert that the document no longer exists
    assert not vector_store_manager.document_exists("4", "test_index")
