from llama_index.embeddings.huggingface_api import \
    HuggingFaceInferenceAPIEmbedding

from .base import BaseEmbeddingModel


class RemoteHuggingFaceEmbedding(BaseEmbeddingModel):
    def __init__(self, model_name: str, api_key: str):
        self.model = HuggingFaceInferenceAPIEmbedding(model_name=model_name, api_key=api_key)

    def get_text_embedding(self, text: str):
        return self.model.get_text_embedding(text)
