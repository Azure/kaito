from fastapi import FastAPI
from fastapi.testclient import TestClient
from inference_api import app

client = TestClient(app)

# Non-Inference Endpoints
def test_read_main():
    response = client.get("/")
    assert response.status_code == 200
    assert response.json() == "Server is running"

def test_health_check():
    response = client.get("/healthz")
    # Assume we have a GPU available and the model & pipeline initialized for testing
    assert response.status_code == 200
    assert response.json() == {"status": "Healthy"}

def test_get_metrics():
    response = client.get("/metrics")
    assert response.status_code == 200
    # Check the structure of the response to ensure GPU metrics are returned
    assert "gpu_info" in response.json()

# Inference Endpoint
def test_text_generation():
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

def test_conversational():
    messages = [
        {"role": "user", "content": "What is your favourite condiment?"},
        {"role": "assistant", "content": "Well, I'm quite partial to a good squeeze of fresh lemon juice. It adds just the right amount of zesty flavour to whatever I'm cooking up in the kitchen!"},
        {"role": "user", "content": "Do you have mayonnaise recipes?"}
    ]
    request_data = {
        "messages": messages,
        "generate_kwargs": {"max_length": 50}  # Example generate_kwargs for conversational
    }
    response = client.post("/chat", json=request_data)
    assert response.status_code == 200
    data = response.json()
    assert "Result" in data
    assert len(data["Result"]) > 0  # Check if the conversation result is not empty

# Invalid tests
def test_invalid_pipeline():
    request_data = {
        "prompt": "This should fail",
        "pipeline": "invalid-pipeline"  # Invalid pipeline type
    }
    response = client.post("/chat", json=request_data)
    assert response.status_code == 400  # Expecting a Bad Request response
    assert "Invalid pipeline type" in response.json().get("detail", "")

def test_missing_prompt():
    request_data = {
        # "prompt" is missing
        "return_full_text": True,
        "clean_up_tokenization_spaces": False
    }
    response = client.post("/chat", json=request_data)
    assert response.status_code == 400  # Expecting a Bad Request response due to missing prompt
    assert "Text generation parameter prompt required" in response.json().get("detail", "")

def test_missing_messages_for_conversation():
    request_data = {
        # "messages" is missing for conversational pipeline
        "pipeline": "conversational"
    }
    response = client.post("/chat", json=request_data)
    assert response.status_code == 400  # Expecting a Bad Request response due to missing messages
    assert "Conversational parameter messages required" in response.json().get("detail", "")


