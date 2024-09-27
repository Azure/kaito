from llama_index.embeddings.huggingface_api import \
    HuggingFaceInferenceAPIEmbedding

from .base import BaseEmbeddingModel


class RemoteHuggingFaceEmbedding(BaseEmbeddingModel):
    def __init__(self, model_name: str, api_key: str):
        self.model = HuggingFaceInferenceAPIEmbedding(model_name=model_name, token=api_key)

    def get_text_embedding(self, text: str):
        """Returns the text embedding for a given input string."""
        return self.model.get_text_embedding(text)
    
    def get_embedding_dimension(self) -> int:
        """Infers the embedding dimension by making a remote call to get the embedding of a dummy text."""
        dummy_input = "This is a dummy sentence."
        embedding = self.get_text_embedding(dummy_input)
        
        # TODO Assume embedding is a 1D array (needs to be tested); return its length (the dimension size)
        return len(embedding)
