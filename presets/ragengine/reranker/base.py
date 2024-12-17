# Copyright (c) Microsoft Corporation.
# Licensed under the MIT license.

from abc import ABC, abstractmethod


class BaseRerankingModel(ABC):
    @abstractmethod
    def rerank(self, text: str):
        """Returns the reranking for a given input string."""
        pass
