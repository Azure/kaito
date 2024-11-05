# Copyright (c) Microsoft Corporation.
# Licensed under the MIT license.

import os
from tempfile import TemporaryDirectory
from unittest.mock import patch
import pytest
from abc import ABC, abstractmethod

from services.vector_store.base import BaseVectorStore
from services.models import Document
from services.embedding.huggingface_local import LocalHuggingFaceEmbedding
from services.config import MODEL_ID, INFERENCE_URL, INFERENCE_ACCESS_SECRET
from services.config import PERSIST_DIR

class BaseVectorStoreTest(ABC):
    """Base class for vector store tests that defines the test structure."""
    
    @pytest.fixture(scope='session')
    def init_embed_manager(self):
        return LocalHuggingFaceEmbedding(MODEL_ID)

    @pytest.fixture
    @abstractmethod
    def vector_store_manager(self, init_embed_manager):
        """Each implementation must provide its own vector store manager."""
        pass

    def test_index_documents(self, vector_store_manager):
        first_doc_text, second_doc_text = "First document", "Second document"
        documents = [
            Document(text=first_doc_text, metadata={"type": "text"}),
            Document(text=second_doc_text, metadata={"type": "text"})
        ]
        
        doc_ids = vector_store_manager.index_documents("test_index", documents)
        
        assert len(doc_ids) == 2
        assert set(doc_ids) == {BaseVectorStore.generate_doc_id(first_doc_text),
                                BaseVectorStore.generate_doc_id(second_doc_text)}

    def test_index_documents_isolation(self, vector_store_manager):
        documents1 = [
            Document(text="First document in index1", metadata={"type": "text"}),
        ]
        documents2 = [
            Document(text="First document in index2", metadata={"type": "text"}),
        ]

        # Index documents in separate indices
        index_name_1, index_name_2 = "index1", "index2"
        vector_store_manager.index_documents(index_name_1, documents1)
        vector_store_manager.index_documents(index_name_2, documents2)

        assert vector_store_manager.list_all_indexed_documents() == {
            'index1': {"87117028123498eb7d757b1507aa3e840c63294f94c27cb5ec83c939dedb32fd":
                           {'hash': '1e64a170be48c45efeaa8667ab35919106da0489ec99a11d0029f2842db133aa',
                            'text': 'First document in index1'}},
            'index2': {"49b198c0e126a99e1975f17b564756c25b4ad691a57eda583e232fd9bee6de91":
                           {'hash': 'a222f875b83ce8b6eb72b3cae278b620de9bcc7c6b73222424d3ce979d1a463b',
                            'text': 'First document in index2'}}
        }

    @patch('requests.post')
    def test_query_documents(self, mock_post, vector_store_manager):
        mock_response = {
            "result": "This is the completion from the API"
        }
        mock_post.return_value.json.return_value = mock_response

        documents = [
            Document(text="First document", metadata={"type": "text"}),
            Document(text="Second document", metadata={"type": "text"})
        ]
        vector_store_manager.index_documents("test_index", documents)

        params = {"temperature": 0.7}
        query_result = vector_store_manager.query("test_index", "First", top_k=1, llm_params=params)

        assert query_result is not None
        assert query_result["response"] == "{'result': 'This is the completion from the API'}"
        assert query_result["source_nodes"][0]["text"] == "First document"
        assert query_result["source_nodes"][0]["score"] == pytest.approx(0.5795239210128784, rel=1e-6)

        mock_post.assert_called_once_with(
            INFERENCE_URL,
            json={"prompt": "Context information is below.\n---------------------\ntype: text\n\nFirst document\n---------------------\nGiven the context information and not prior knowledge, answer the query.\nQuery: First\nAnswer: ", "formatted": True, 'temperature': 0.7},
            headers={"Authorization": f"Bearer {INFERENCE_ACCESS_SECRET}"}
        )

    def test_add_document(self, vector_store_manager):
        documents = [Document(text="Third document", metadata={"type": "text"})]
        vector_store_manager.index_documents("test_index", documents)

        new_document = [Document(text="Fourth document", metadata={"type": "text"})]
        vector_store_manager.index_documents("test_index", new_document)

        assert vector_store_manager.document_exists("test_index", new_document[0],
                                                    BaseVectorStore.generate_doc_id("Fourth document"))

    def test_persist_index_1(self, vector_store_manager):
        documents = [Document(text="Test document", metadata={"type": "text"})]
        vector_store_manager.index_documents("test_index", documents)
        vector_store_manager._persist("test_index")
        assert os.path.exists(PERSIST_DIR)

    def test_persist_index_2(self, vector_store_manager):
        documents = [Document(text="Test document", metadata={"type": "text"})]
        vector_store_manager.index_documents("test_index", documents)

        documents = [Document(text="Another Test document", metadata={"type": "text"})]
        vector_store_manager.index_documents("another_test_index", documents)

        vector_store_manager._persist_all()
        assert os.path.exists(PERSIST_DIR)
