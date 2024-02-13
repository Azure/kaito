# Copyright (c) Microsoft Corporation.
# Licensed under the MIT license.
import sys
from pathlib import Path

# Get the parent directory of the current file
parent_dir = str(Path(__file__).resolve().parent.parent)
# Add the parent directory to sys.path
sys.path.append(parent_dir)

import argparse
from unittest.mock import patch

from fastapi.testclient import TestClient

# Parse the command-line arguments
parser = argparse.ArgumentParser()
parser.add_argument("--pipeline", required=True, help="Pipeline type")
parser.add_argument("--pretrained_model_name_or_path", required=True, help="Model path")
parser.add_argument("--allow_remote_files", default=True, help="Allow models to be downloaded for tests")
args = parser.parse_args()
pipeline_type = args.pipeline

try:
    from inference_api import ModelConfig, app
except ValueError as e:
    if pipeline_type not in {"text-generation", "conversational"}:
        # Pipeline is invalid, handle and exit
        print(f"Correctly caught invalid pipeline during import")
        sys.exit(0)
    else:
        raise
except Exception as e:
    # For all other exceptions, re-raise
    raise
    
def run_tests(): 
    client = TestClient(app)
    test_read_main(client)
    test_health_check(client)
    test_get_metrics(client)
    test_get_metrics_no_gpus(client)
    # Pipeline must be valid to pass import
    if pipeline_type == "text-generation":
        test_text_generation(client)
        test_missing_prompt(client)
    elif pipeline_type == "conversational":
        test_conversational(client)
        test_missing_messages_for_conversation(client)
        
def test_read_main(client):
    response = client.get("/")
    server_msg, status_code = response.json()
    assert server_msg == "Server is running"
    assert status_code == 200
    
def test_health_check(client):
    response = client.get("/healthz")    
    # Assume we have a GPU available and the model & pipeline initialized for testing
    assert response.status_code == 200
    assert response.json() == {"status": "Healthy"}

def test_get_metrics(client):
    response = client.get("/metrics")
    assert response.status_code == 200
    # Check the structure of the response to ensure GPU metrics are returned
    assert "gpu_info" in response.json()

def test_get_metrics_no_gpus(client):
    with patch('GPUtil.getGPUs', return_value=[]) as mock_getGPUs:
        response = client.get("/metrics")
        assert response.status_code == 200
        assert response.json()["gpu_info"] == []  # Expecting an empty list

def test_text_generation(client):
    request_data = {
        "prompt": "Hello, world!",
        "return_full_text": True,
        "clean_up_tokenization_spaces": False,
        "generate_kwargs": {"max_length": 50, "min_length": 10}  # Example generate_kwargs
    }
    response = client.post("/chat", json=request_data)
    assert response.status_code == 200
    data = response.json()
    assert "Result" in data
    assert len(data["Result"]) > 0  # Check if the result text is not empty

def test_missing_prompt(client):
    request_data = {
        # "prompt" is missing
        "return_full_text": True,
        "clean_up_tokenization_spaces": False,
        "generate_kwargs": {"max_length": 50}
    }
    response = client.post("/chat", json=request_data)
    assert response.status_code == 400  # Expecting a Bad Request response due to missing prompt
    assert "Text generation parameter prompt required" in response.json().get("detail", "")
    
def test_conversational(client):
    messages = [
        {"role": "user", "content": "What is your favourite condiment?"},
        {"role": "assistant", "content": "Well, Im quite partial to a good squeeze of fresh lemon juice. It adds just the right amount of zesty flavour to whatever Im cooking up in the kitchen!"},
        {"role": "user", "content": "Do you have mayonnaise recipes?"}
    ]
    request_data = {
        "messages": messages,
        "generate_kwargs": {"max_new_tokens": 1000, "do_sample": True}
    }
    response = client.post("/chat", json=request_data)

    assert response.status_code == 200
    data = response.json()
    assert "Result" in data
    assert len(data["Result"]) > 0  # Check if the conversation result is not empty

def test_missing_messages_for_conversation(client):
    request_data = {
        # "messages" is missing for conversational pipeline
    }
    response = client.post("/chat", json=request_data)
    assert response.status_code == 400  # Expecting a Bad Request response due to missing messages
    assert "Conversational parameter messages required" in response.json().get("detail", "")

if __name__ == "__main__":
    run_tests()