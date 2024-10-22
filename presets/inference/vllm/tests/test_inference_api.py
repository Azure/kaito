import sys
import os
import subprocess
import time
import socket
from pathlib import Path

import pytest
import requests

# Get the parent directory of the current file
parent_dir = str(Path(__file__).resolve().parent.parent)
# Add the parent directory to sys.path
sys.path.append(parent_dir)

TEST_MODEL = "facebook/opt-125m"
CHAT_TEMPLATE = ("{{ bos_token }}{% for message in messages %}{% if (message['role'] == 'user') %}"
    "{{'<|user|>' + '\n' + message['content'] + '<|end|>' + '\n' + '<|assistant|>' + '\n'}}"
    "{% elif (message['role'] == 'assistant') %}{{message['content'] + '<|end|>' + '\n'}}{% endif %}{% endfor %}")

@pytest.fixture(scope="module", autouse=True)
def setup_server(request):
    print("\nDoing setup")
    port = find_available_port()
    global TEST_PORT
    TEST_PORT = port

    args = [
        "python3",
        os.path.join(parent_dir, "inference_api.py"),
        "--model", TEST_MODEL,
        "--chat-template", CHAT_TEMPLATE,
        "--port", str(TEST_PORT)
    ]
    print(f"Starting server on port {TEST_PORT}")
    process = subprocess.Popen(args, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    if not is_port_open("localhost", TEST_PORT):
        fin()
        pytest.fail("failed to launch vllm server")

    def fin():
        process.terminate()
        process.wait()
        print ("\nDoing teardown")
    request.addfinalizer(fin)

def is_port_open(host, port, timeout=60):
    start_time = time.time()
    while time.time() - start_time < timeout:
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:
            sock.settimeout(1)  # Set a short timeout for each connection attempt
            result = sock.connect_ex((host, port))
            print(">>> waiting for server to start")
            if result == 0:
                print(f">>> server started in {int(time.time() - start_time)} seconds")
                return True
            time.sleep(1)  # Wait a bit before retrying
    return False

def find_available_port(start_port=5000, end_port=8000):
    for port in range(start_port, end_port + 1):
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
            if s.connect_ex(('localhost', port)) != 0:
                return port
    raise RuntimeError('No available ports found')

def test_completions_api(setup_server):
    request_data = {
        "model": TEST_MODEL,
        "prompt": "Say this is a test",
        "max_tokens": 7,
        "temperature": 0.5,
        "n": 2
    }

    response = requests.post(f"http://127.0.0.1:{TEST_PORT}/v1/completions", json=request_data)
    data = response.json()
    assert "choices" in data, "The response should contain a 'choices' key"
    assert len(data["choices"]) == 2, "The response should contain two completion"

    for choice in data["choices"]:
        assert "text" in choice, "Each choice should contain a 'text' key"
        assert len(choice["text"]) > 0, "The completion text should not be empty"

def test_chat_completions_api(setup_server):
    request_data = {
        "model": TEST_MODEL,
        "messages": [
            {"role": "user", "content": "Hello!"},
            {"role": "assistant", "content": "Hi there! How can I help you today?"}
        ],
        "max_tokens": 7,
        "temperature": 0.5,
        "n": 2
    }

    response = requests.post(f"http://127.0.0.1:{TEST_PORT}/v1/chat/completions", json=request_data)
    data = response.json()

    assert "choices" in data, "The response should contain a 'choices' key"
    assert len(data["choices"]) == 2, "The response should contain two completion"

    for choice in data["choices"]:
        assert "message" in choice, "Each choice should contain a 'message' key"
        assert "content" in choice["message"], "Each message should contain a 'content' key"
        assert len(choice["message"]["content"]) > 0, "The completion text should not be empty"