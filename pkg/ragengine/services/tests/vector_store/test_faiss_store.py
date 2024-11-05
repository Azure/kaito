# Copyright (c) Microsoft Corporation.
# Licensed under the MIT license.

import pytest
import os

from tempfile import TemporaryDirectory
from services.tests.vector_store.test_base_store import BaseVectorStoreTest
from services.vector_store.faiss_store import FaissVectorStoreHandler

class TestFaissVectorStore(BaseVectorStoreTest):
    """Test implementation for FAISS vector store."""
    
    @pytest.fixture
    def vector_store_manager(self, init_embed_manager):
        with TemporaryDirectory() as temp_dir:
            print(f"Saving temporary test storage at: {temp_dir}")
            os.environ['PERSIST_DIR'] = temp_dir
            yield FaissVectorStoreHandler(init_embed_manager)
