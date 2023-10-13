from fastapi import FastAPI, HTTPException
import uvicorn
from pydantic import BaseModel
from typing import Optional
import threading

from llama import Llama
import torch
import sys
import signal
import os
import torch.distributed as dist
import argparse

# Setup argparse
parser = argparse.ArgumentParser(description="Llama API server.")
parser.add_argument("--ckpt_dir", default="weights/", help="Checkpoint directory.")
parser.add_argument("--tokenizer_path", default="tokenizer.model", help="Path to the tokenizer model.")
parser.add_argument("--max_seq_len", type=int, default=128, help="Maximum sequence length.")
parser.add_argument("--max_batch_size", type=int, default=4, help="Maximum batch size.")
parser.add_argument("--model_parallel_size", type=int, default=int(os.environ.get("WORLD_SIZE", 1)), help="Model parallel size.")
args = parser.parse_args()

should_shutdown = False

def build_generator(params):
    """Build Llama generator from provided parameters."""
    return Llama.build(**params)

def broadcast_for_shutdown():
    """Broadcasts shutdown command to worker processes."""
    dist.broadcast_object_list(["shutdown", None, None], src=0)

def broadcast_for_generation(input_string, max_gen_len, temperature, top_p):
    """Broadcasts generation parameters to worker processes."""
    dist.broadcast_object_list(["generate", input_string, {
        'max_gen_len': max_gen_len,
        'temperature': temperature,
        'top_p': top_p
    }], src=0)

def shutdown_server():
    """Shut down the server."""
    os.kill(os.getpid(), signal.SIGINT)

# Default values for the generator
gen_params = {
    'ckpt_dir': args.ckpt_dir,
    'tokenizer_path': args.tokenizer_path,
    'max_seq_len': args.max_seq_len,
    'max_batch_size': args.max_batch_size,
    'model_parallel_size': args.model_parallel_size,
}

generator = build_generator(gen_params)

def setup_main_routes(): 
    @app_main.get('/')
    def home():
        return "Server is running", 200

    @app_main.get("/healthz")
    def health_check():
        if not torch.cuda.is_available():
            raise HTTPException(status_code=500, detail="No GPU available")
        if not generator:
            raise HTTPException(status_code=500, detail="Llama model not initialized")
        return {"status": "Healthy"}

    @app_main.post("/shutdown")
    def shutdown():
        """Shutdown the server and worker processes."""
        global should_shutdown
        should_shutdown = True
        if dist.get_world_size() > 1:
            broadcast_for_shutdown()
        shutdown_server()
        return {}

    class ChatParameters(BaseModel):
        input_data: dict
        parameters: Optional[dict] = None

    @app_main.post("/chat")
    def chat_completion(params: ChatParameters):
        input_data = params.input_data
        if not input_data:
            raise HTTPException(status_code=400, detail="Input data is required")
        
        input_string = input_data.get("input_string")
        if not input_string:
            raise HTTPException(status_code=400, detail="Input string is required")

        parameters = params.parameters if params.parameters else {}
        max_gen_len = parameters.get('max_gen_len', None)
        temperature = parameters.get('temperature', 0.6)
        top_p = parameters.get('top_p', 0.9)

        if dist.get_world_size() > 1:
            # Broadcast generation params to worker processes
            broadcast_for_generation(input_string, max_gen_len, temperature, top_p)

        # Master's own generation
        try:
            results = generator.chat_completion(
                input_string,
                max_gen_len=max_gen_len,
                temperature=temperature,
                top_p=top_p,
            )
        except Exception as e:
            raise HTTPException(status_code=400, detail="Request Failed: " + str(e))

        if len(results) == 0:
            raise HTTPException(status_code=404, detail="No results")
        
        response_data = []
        for dialog, result in zip(input_string, results):
            conversation = []
            for msg in dialog:
                print(f"{msg['role'].capitalize()}: {msg['content']}\n")
                conversation.append({
                    "role": msg['role'].capitalize(),
                    "content": msg['content']
                })
            print(
                f"> {result['generation']['role'].capitalize()}: {result['generation']['content']}"
            )
            conversation.append({
                "role": result['generation']['role'].capitalize(),
                "content": result['generation']['content']
            })
            response_data.append(conversation)
            print("\n==================================\n")

        return {"results": response_data}

def setup_worker_routes():
    @app_worker.get("/healthz")
    def health_check():
        if not torch.cuda.is_available():
            raise HTTPException(status_code=500, detail="No GPU available")
        if not generator:
            raise HTTPException(status_code=500, detail="Llama model not initialized")
        return {"status": "Healthy"}

def start_worker_server():
    uvicorn.run(app=app_worker, host='0.0.0.0', port=5000)
    print(f"Worker {dist.get_rank()} HTTP health server started at port 5000")

def worker_listen_tasks(): 
    while True:
        worker_num = dist.get_rank()
        print(f"Worker {worker_num} ready to recieve next command")
        config = [None] * 3  # Command and its associated data
        dist.broadcast_object_list(config, src=0)
        command = config[0]

        if command == "generate":
            try:
                input_string = config[1]
                parameters = config[2]
                generator.chat_completion(
                    input_string,
                    max_gen_len=parameters.get('max_gen_len', None),
                    temperature=parameters.get('temperature', 0.6),
                    top_p=parameters.get('top_p', 0.9)
                )
                print(f"Worker {worker_num} completed generation")              
            except Exception as e:
                print(f"Error in generation: {str(e)}")
        elif command == "shutdown":
            print(f"Worker {worker_num} shutting down")
            sys.exit(0)


if __name__ == "__main__":
    # Fetch the LOCAL_RANK environment variable to determine the rank of this process
    # on the current node (machine).
    local_rank = int(os.environ.get("LOCAL_RANK")) 

    # dist.get_rank() provides the global rank across all nodes. 
    # The following code is run by the globally ranked process 0.
    if dist.get_rank() == 0:
        # This is the main server that handles the main logic of our application.
        app_main = FastAPI()
        setup_main_routes()
        uvicorn.run(app=app_main, host='0.0.0.0', port=5000)  # Use the app_main instance
    else:
        # This code is executed by all processes that aren't the globally ranked 0.
        # This includes processes on the main node as well as on other nodes.

        # Uncomment to enable worker logs
        # sys.stdout = sys.__stdout__

        # If the current process is the locally ranked 0 (i.e., the primary process)
        # on its node, then it starts a worker server that exposes a health check endpoint.
        if local_rank == 0:
            app_worker = FastAPI()
            setup_worker_routes()
             
            # Start the worker server in a separate thread. This worker server will
            # provide a healthz endpoint for monitoring the health of the node.
            server_thread = threading.Thread(target=start_worker_server, daemon=True)
            server_thread.start()

        # Regardless of local rank, all non-globally-0-ranked processes will listen
        # for tasks (like chat completion) from the main server.
        worker_listen_tasks()
