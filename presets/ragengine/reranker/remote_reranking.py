# Copyright (c) Microsoft Corporation.
# Licensed under the MIT license.

import requests
import json
from .base import BaseRerankingModel


class RemoteRerankingModel(BaseRerankingModel):
    def __init__(self, model_url: str, api_key: str):
        """
        Initialize the RemoteEmbeddingModel.

        Args:
            model_url (str): The URL of the embedding model API endpoint.
            api_key (str): The API key for accessing the API.
        """
        self.model_url = model_url
        self.api_key = api_key

    def rerank(self, text: str):
        """Returns the reranking for a given input string."""
        headers = {
            "Authorization": f"Bearer {self.api_key}",
            "Content-Type": "application/json"
        }
        payload = {
            "inputs": text
        }

        try:
            response = requests.post(self.model_url, headers=headers, data=json.dumps(payload))
            response.raise_for_status()  # Raise an HTTPError for bad responses
            reranking = response.json()  # Assumes the API returns JSON
            if isinstance(reranking, list):
                return reranking
            else:
                raise ValueError("Unexpected response format. Expected a list.")
        except requests.exceptions.RequestException as e:
            raise RuntimeError(f"Failed to get embedding from remote model: {e}")
