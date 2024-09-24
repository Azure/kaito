from llama_index.core import Settings
from llama_index.llms.huggingface_api import HuggingFaceInferenceAPI

remote_llm_api = HuggingFaceInferenceAPI(
    model_name="HuggingFaceH4/zephyr-7b-alpha"
)

Settings.llm = remote_llm_api

import logging

import chromadb
from IPython.display import Markdown, display
from llama_index.core import (SimpleDirectoryReader, StorageContext,
                              VectorStoreIndex)
from llama_index.embeddings.huggingface import HuggingFaceEmbedding
from llama_index.vector_stores.chroma import ChromaVectorStore

# Enable DEBUG logging for ChromaDB
logging.basicConfig(level=logging.DEBUG)

# create ChromaDB client and a new collection
chroma_client = chromadb.EphemeralClient()
chroma_collection = chroma_client.create_collection("quickstart")

# define embedding function
embed_model = HuggingFaceEmbedding(model_name="BAAI/bge-base-en-v1.5")

# load documents from directory
documents = SimpleDirectoryReader("./data/paul_graham/").load_data()

# set up ChromaVectorStore and load in data
vector_store = ChromaVectorStore(chroma_collection=chroma_collection)
storage_context = StorageContext.from_defaults(vector_store=vector_store)
index = VectorStoreIndex.from_documents(
    documents, storage_context=storage_context, embed_model=embed_model
)

# Log collection contents before querying
logging.debug("Documents in ChromaDB collection before querying:")
all_documents = chroma_collection.get(include=["documents"])
logging.debug(all_documents["documents"])

# Query Data
query_engine = index.as_query_engine()
response = query_engine.query("What did the author do growing up?")
display(Markdown(f"{response}"))

# Log collection contents after querying
logging.debug("Documents in ChromaDB collection after querying:")
all_documents_after_query = chroma_collection.get(include=["documents"])
logging.debug(all_documents_after_query["documents"])

# Log embeddings stored in ChromaDB
logging.debug("Embeddings stored in ChromaDB:")
all_embeddings = chroma_collection.get(include=["embeddings"])
logging.debug(all_embeddings["embeddings"])

# Log metadata stored in ChromaDB
logging.debug("Metadata stored in ChromaDB:")
all_metadata = chroma_collection.get(include=["metadatas"])
logging.debug(all_metadata["metadatas"])
