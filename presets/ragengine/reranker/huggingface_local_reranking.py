# Copyright (c) Microsoft Corporation.
# Licensed under the MIT license.

from llama_index.embeddings.huggingface import HuggingFaceEmbedding

from .base import BaseRerankingModel


class LocalHuggingFaceReranking(BaseRerankingModel):
    def __init__(self, model_name: str):
        self.model = HuggingFaceEmbedding(model_name=model_name) # TODO: Ensure/test loads on GPU (when available)

    def rerank(self, text: str):
        """Returns the reranking for a given input string."""
        pass