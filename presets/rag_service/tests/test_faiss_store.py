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
        print(f"Saving Temporary Test Storage at: {temp_dir}")
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

def test_index_documents_isolation(vector_store_manager):
    doc_1_id, doc_2_id = "1", "2"
    documents1 = [
        Document(doc_id=doc_1_id, text="First document in index1", metadata={"type": "text"}),
    ]
    documents2 = [
        Document(doc_id=doc_2_id, text="First document in index2", metadata={"type": "text"}),
    ]

    # Index documents in separate indices
    index_name_1, index_name_2 = "index1", "index2"
    vector_store_manager.index_documents(documents1, index_name=index_name_1)
    vector_store_manager.index_documents(documents2, index_name=index_name_2)

    # Ensure documents are correctly persisted and separated by index
    doc_1 = vector_store_manager.get_document(doc_1_id, index_name=index_name_1)
    assert doc_1 and doc_1.node_ids # Ensure documents were created

    doc_2 = vector_store_manager.get_document(doc_2_id, index_name=index_name_2)
    assert doc_2 and doc_2.node_ids # Ensure documents were created

    # Ensure that the documents do not mix between indices
    assert vector_store_manager.get_document(doc_1_id, index_name=index_name_2) is None, f"Document {doc_1_id} should not exist in {index_name_2}"
    assert vector_store_manager.get_document(doc_2_id, index_name=index_name_1) is None, f"Document {doc_2_id} should not exist in {index_name_1}"


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
