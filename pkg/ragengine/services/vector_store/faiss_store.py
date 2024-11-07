# Copyright (c) Microsoft Corporation.
# Licensed under the MIT license.

from typing import List

import faiss
from llama_index.vector_stores.faiss import FaissVectorStore
from services.models import Document
from .base import BaseVectorStore


class FaissVectorStoreHandler(BaseVectorStore):
    def __init__(self, embedding_manager):
        super().__init__(embedding_manager)
        self.dimension = self.embedding_manager.get_embedding_dimension()

    def _create_new_index(self, index_name: str, documents: List[Document]) -> List[str]:
        faiss_index = faiss.IndexFlatL2(self.dimension)
        vector_store = FaissVectorStore(faiss_index=faiss_index)
        return self._create_index_common(index_name, documents, vector_store)
