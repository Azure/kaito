from llama_index.embeddings.huggingface import HuggingFaceEmbedding

from .base import BaseEmbeddingModel


class LocalHuggingFaceEmbedding(BaseEmbeddingModel):
    def __init__(self, model_name: str):
        self.model = HuggingFaceEmbedding(model_name=model_name)

    def get_text_embedding(self, text: str):
        return self.model.get_text_embedding(text)
