import os
from tempfile import TemporaryDirectory
from unittest.mock import patch

import pytest
from vector_store.faiss_store import FaissVectorStoreManager
from models import Document
from embedding.huggingface_local import LocalHuggingFaceEmbedding
from config import MODEL_ID, INFERENCE_URL, INFERENCE_ACCESS_SECRET

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

@patch('requests.post')
def test_query_documents(mock_post, vector_store_manager):
    # Define Mock Response for Custom Inference API
    mock_response = {
        "result": "This is the completion from the API"
    }

    mock_post.return_value.json.return_value = mock_response

    # Add documents to index
    documents = [
        Document(doc_id="1", text="First document", metadata={"type": "text"}),
        Document(doc_id="2", text="Second document", metadata={"type": "text"})
    ]
    vector_store_manager.index_documents(documents, index_name="test_index")

    # Mock query and results
    query_result = vector_store_manager.query("First", top_k=1, index_name="test_index")

    assert query_result is not None
    assert query_result.response == "This is the completion from the API"

    mock_post.assert_called_once_with(
        INFERENCE_URL,
        json={"prompt": "Context information is below.\n---------------------\ntype: text\n\nFirst document\n---------------------\nGiven the context information and not prior knowledge, answer the query.\nQuery: First\nAnswer: ", "formatted": True},
        headers={"Authorization": f"Bearer {INFERENCE_ACCESS_SECRET}"}
    )

def test_add_and_delete_document(vector_store_manager, capsys):
    documents = [Document(doc_id="3", text="Third document", metadata={"type": "text"})]
    vector_store_manager.index_documents(documents, index_name="test_index")

    # Add a document to the existing index
    new_document = Document(doc_id="4", text="Fourth document", metadata={"type": "text"})
    vector_store_manager.add_document(new_document, index_name="test_index")

    # Assert that the document exists
    assert vector_store_manager.document_exists("4", "test_index")

    # Delete the document - it should handle the NotImplementedError and not raise an exception
    vector_store_manager.delete_document("4", "test_index")

    # Capture the printed output (if any)
    captured = capsys.readouterr()

    # Check if the expected message about NotImplementedError was printed
    assert "Delete not yet implemented for Faiss index. Skipping document 4." in captured.out

    # Assert that the document still exists (since deletion wasn't implemented)
    assert vector_store_manager.document_exists("4", "test_index")


def test_update_document_not_implemented(vector_store_manager, capsys):
    """Test that updating a document raises a NotImplementedError and is handled properly."""
    # Add a document to the index
    documents = [Document(doc_id="1", text="First document", metadata={"type": "text"})]
    vector_store_manager.index_documents(documents, index_name="test_index")

    # Attempt to update the existing document
    updated_document = Document(doc_id="1", text="Updated first document", metadata={"type": "text"})
    vector_store_manager.update_document(updated_document, index_name="test_index")

    # Capture the printed output (if any)
    captured = capsys.readouterr()

    # Check if the NotImplementedError message was printed
    assert "Update is equivalent to deleting the document and then inserting it again." in captured.out
    assert f"Update not yet implemented for Faiss index. Skipping document {updated_document.doc_id}." in captured.out

    # Ensure the document remains unchanged
    original_doc = vector_store_manager.get_document("1", index_name="test_index")
    assert original_doc is not None


def test_refresh_unchanged_documents(vector_store_manager, capsys):
    """Test that refreshing documents does nothing on unchanged documents."""
    # Add documents to the index
    documents = [Document(doc_id="1", text="First document", metadata={"type": "text"}),
                 Document(doc_id="2", text="Second document", metadata={"type": "text"})]
    vector_store_manager.index_documents(documents, index_name="test_index")

    refresh_results = vector_store_manager.refresh_documents(documents, index_name="test_index")

    # Capture the printed output (if any)
    captured = capsys.readouterr()
    assert captured.out == ""
    assert refresh_results == [False, False]

def test_refresh_new_documents(vector_store_manager):
    """Test that refreshing new documents creates them."""
    vector_store_manager.index_documents([], index_name="test_index")

    # Add a document to the index
    documents = [Document(doc_id="1", text="First document", metadata={"type": "text"}),
                 Document(doc_id="2", text="Second document", metadata={"type": "text"})]

    refresh_results = vector_store_manager.refresh_documents(documents, index_name="test_index")

    inserted_documents = vector_store_manager.list_documents(index_name="test_index")

    assert len(inserted_documents) == len(documents)
    assert inserted_documents.keys() == {"1", "2"}
    assert refresh_results == [True, True]

def test_refresh_existing_documents(vector_store_manager, capsys):
    """Test that refreshing existing documents prints error."""
    original_documents = [Document(doc_id="1", text="First document", metadata={"type": "text"})]
    vector_store_manager.index_documents(original_documents, index_name="test_index")

    new_documents = [Document(doc_id="1", text="Updated document", metadata={"type": "text"}),
                     Document(doc_id="2", text="Second document", metadata={"type": "text"})]

    refresh_results = vector_store_manager.refresh_documents(new_documents, index_name="test_index")

    captured = capsys.readouterr()

    # Check if the NotImplementedError message was printed
    assert "Refresh not yet fully implemented for index" in captured.out
    assert not refresh_results
