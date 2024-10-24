from typing import List
import os
from ragengine.models import Document

import pymongo
from llama_index.vector_stores.azurecosmosmongo import (
    AzureCosmosDBMongoDBVectorSearch,
)

from .base import BaseVectorStore

class AzureCosmosDBMongoDBVectorStoreHandler(BaseVectorStore):
    def __init__(self, embedding_manager):
        super().__init__(embedding_manager)
        self.connection_string = os.environ.get("AZURE_COSMOSDB_MONGODB_URI")
        self.mongodb_client = pymongo.MongoClient(self.connection_string)

    def _create_new_index(self, index_name: str, documents: List[Document]) -> List[str]:
        vector_store = AzureCosmosDBMongoDBVectorSearch(
            mongodb_client=self.mongodb_client,
            db_name="kaito_ragengine",
            collection_name=index_name,
        )
        return self._create_index_common(index_name, documents, vector_store)
