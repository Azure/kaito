import importlib
import sys
from pathlib import Path
from unittest.mock import patch

import pytest
import torch
from fastapi.testclient import TestClient

# Get the parent directory of the current file
parent_dir = str(Path(__file__).resolve().parent.parent)
# Add the parent directory to sys.path
sys.path.append(parent_dir)

@pytest.fixture(params=[
    {"pipeline": "text-generation", "model_path": "stanford-crfm/alias-gpt2-small-x21"},
    {"pipeline": "conversational", "model_path": "stanford-crfm/alias-gpt2-small-x21"},
])
def configured_app(request):
    original_argv = sys.argv.copy()
    # Use request.param to set correct test arguments for each configuration
    test_args = [
        'program_name',
        '--pipeline', request.param['pipeline'],
        '--pretrained_model_name_or_path', request.param['model_path'],
        '--allow_remote_files', 'True'
    ]
    sys.argv = test_args

    import inference_api
    importlib.reload(inference_api) # Reload to prevent module caching
    from inference_api import app

    # Attach the request params to the app instance for access in tests
    app.test_config = request.param
    yield app

    sys.argv = original_argv

def test_conversational(configured_app):
    if configured_app.test_config['pipeline'] != 'conversational':
        pytest.skip("Skipping non-conversational tests")
    client = TestClient(configured_app)
    messages = [
        {"role": "user", "content": "What is your favourite condiment?"},
        {"role": "assistant", "content": "Well, Im quite partial to a good squeeze of fresh lemon juice. It adds just the right amount of zesty flavour to whatever Im cooking up in the kitchen!"},
        {"role": "user", "content": "Do you have mayonnaise recipes?"}
    ]
    request_data = {
        "messages": messages,
        "generate_kwargs": {"max_new_tokens": 20, "do_sample": True}
    }
    response = client.post("/chat", json=request_data)

    assert response.status_code == 200
    data = response.json()
    assert "Result" in data
    assert len(data["Result"]) > 0  # Check if the conversation result is not empty

def test_missing_messages_for_conversation(configured_app):
    if configured_app.test_config['pipeline'] != 'conversational':
        pytest.skip("Skipping non-conversational tests")
    client = TestClient(configured_app)
    request_data = {
        # "messages" is missing for conversational pipeline
    }
    response = client.post("/chat", json=request_data)
    assert response.status_code == 400  # Expecting a Bad Request response due to missing messages
    assert "Conversational parameter messages required" in response.json().get("detail", "")

def test_text_generation(configured_app):
    if configured_app.test_config['pipeline'] != 'text-generation':
        pytest.skip("Skipping non-text-generation tests")
    client = TestClient(configured_app)
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

def test_missing_prompt(configured_app):
    if configured_app.test_config['pipeline'] != 'text-generation':
        pytest.skip("Skipping non-text-generation tests")
    client = TestClient(configured_app)
    request_data = {
        # "prompt" is missing
        "return_full_text": True,
        "clean_up_tokenization_spaces": False,
        "generate_kwargs": {"max_length": 50}
    }
    response = client.post("/chat", json=request_data)
    assert response.status_code == 400  # Expecting a Bad Request response due to missing prompt
    assert "Text generation parameter prompt required" in response.json().get("detail", "")

def test_read_main(configured_app):
    client = TestClient(configured_app)
    response = client.get("/")
    server_msg, status_code = response.json()
    assert server_msg == "Server is running"
    assert status_code == 200

def test_health_check(configured_app):
    device = "GPU" if torch.cuda.is_available() else "CPU"
    if device != "GPU":
        pytest.skip("Skipping healthz endpoint check - running on CPU")
    client = TestClient(configured_app)
    response = client.get("/healthz")
    # Assuming we have a GPU available
    assert response.status_code == 200
    assert response.json() == {"status": "Healthy"}

def test_get_metrics(configured_app):
    client = TestClient(configured_app)
    response = client.get("/metrics")
    assert response.status_code == 200
    assert "gpu_info" in response.json()

def test_get_metrics_no_gpus(configured_app):
    client = TestClient(configured_app)
    with patch('GPUtil.getGPUs', return_value=[]) as mock_getGPUs:
        response = client.get("/metrics")
        assert response.status_code == 200
        assert response.json()["gpu_info"] == []

def test_default_generation_params(configured_app):
    if configured_app.test_config['pipeline'] != 'text-generation':
        pytest.skip("Skipping non-text-generation tests")
    
    client = TestClient(configured_app)

    request_data = {
        "prompt": "Test default params",
        "return_full_text": True,
        "clean_up_tokenization_spaces": False
        # Note: generate_kwargs is not provided, so defaults should be used
    }

    with patch('inference_api.pipeline') as mock_pipeline:
        mock_pipeline.return_value = [{"generated_text": "Mocked response"}]  # Mock the response of the pipeline function
        
        response = client.post("/chat", json=request_data)
        
        assert response.status_code == 200
        data = response.json()
        assert "Result" in data
        assert len(data["Result"]) > 0
        
        # Check the default args
        _, kwargs = mock_pipeline.call_args
        assert kwargs['max_length'] == 200
        assert kwargs['min_length'] == 0
        assert kwargs['do_sample'] is False
        assert kwargs['temperature'] == 1.0
        assert kwargs['top_k'] == 50
        assert kwargs['top_p'] == 1
        assert kwargs['typical_p'] == 1
        assert kwargs['repetition_penalty'] == 1
        assert kwargs['num_beams'] == 1
        assert kwargs['early_stopping'] is False