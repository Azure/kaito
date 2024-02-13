# Copyright (c) Microsoft Corporation.
# Licensed under the MIT license.
import argparse
import functools
import multiprocessing
import multiprocessing.pool
import os
import signal
import sys
import threading
from typing import Optional

import GPUtil
import torch
import torch.distributed as dist
import uvicorn
from fastapi import FastAPI, HTTPException
from llama import Llama
from pydantic import BaseModel

# Setup argparse
parser = argparse.ArgumentParser(description="Llama API server.")
parser.add_argument("--ckpt_dir", default="weights/", help="Checkpoint directory.")
parser.add_argument("--tokenizer_path", default="weights/tokenizer.model", help="Path to the tokenizer model.")
parser.add_argument("--max_seq_len", type=int, default=128, help="Maximum sequence length.")
parser.add_argument("--max_batch_size", type=int, default=4, help="Maximum batch size.")
parser.add_argument("--model_parallel_size", type=int, default=int(os.environ.get("WORLD_SIZE", 1)), help="Model parallel size.")
args = parser.parse_args()

should_shutdown = False

def timeout(max_timeout):
    """Timeout decorator, parameter in seconds."""
    def timeout_decorator(item):
        """Wrap the original function."""
        @functools.wraps(item)
        def func_wrapper(*args, **kwargs):
            """Closure for function."""
            with multiprocessing.pool.ThreadPool(processes=1) as pool:
                async_result = pool.apply_async(item, args, kwargs)
                # raises a TimeoutError if execution exceeds max_timeout
                return async_result.get(max_timeout)
        return func_wrapper
    return timeout_decorator

def build_generator(params):
    """Build Llama generator from provided parameters."""
    return Llama.build(**params)

def broadcast_for_shutdown():
    """Broadcasts shutdown command to worker processes."""
    dist.broadcast_object_list(["shutdown", None, None], src=0)

def broadcast_for_text_generation(prompts, max_gen_len, temperature, top_p):
    """Broadcasts generation parameters to worker processes."""
    dist.broadcast_object_list(["text_generate", prompts, {
        'max_gen_len': max_gen_len,
        'temperature': temperature,
        'top_p': top_p
    }], src=0)

@timeout(180.0)
def master_inference(prompts, max_gen_len, temperature, top_p):
    if dist.get_world_size() > 1:
        try:
            # Broadcast generation params to worker processes
            broadcast_for_text_generation(prompts, max_gen_len, temperature, top_p)
        except Exception as e:
            print("Error in broadcast_for_text_generation:", str(e))
            raise

    # Master's own generation
    try:
        return generator.text_completion(
            prompts,
            max_gen_len=max_gen_len,
            temperature=temperature,
            top_p=top_p,
        )
    except Exception as e:
        print("Error in text_completion:", str(e))
        raise

def shutdown_server():
    """Shut down the server."""
    os.killpg(os.getpgrp(), signal.SIGTERM)

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

    class GenerationParameters(BaseModel):
        prompts: list
        parameters: Optional[dict] = None

    @app_main.post("/generate")
    def generate_text(params: GenerationParameters):
        prompts = params.prompts
        # Check if the prompts are provided
        if not prompts or not isinstance(prompts, list):
            raise HTTPException(status_code=400, detail="Prompts are required and should be an array")

        parameters = params.parameters if params.parameters else {}
        max_gen_len = parameters.get('max_gen_len', None)
        temperature = parameters.get('temperature', 0.6)
        top_p = parameters.get('top_p', 0.9)

        try: 
            results = master_inference(prompts, max_gen_len, temperature, top_p)
        except Exception as e: 
            exception_type = type(e).__name__
            if exception_type == "TimeoutError": 
                print("Master Inference Failed - TimeoutError", e)
                raise HTTPException(status_code=408, detail="Request Timed Out")
            raise HTTPException(status_code=400, detail="Request Failed: " + str(e))

        if len(results) == 0:
            raise HTTPException(status_code=404, detail="No results")

        response_data = []
        for prompt, result in zip(prompts, results):
            print(prompt)
            print(f"> {result['generation']}")
            print("\n==================================\n")
            entry = {
                "prompt": prompt,
                "response": result['generation']
            }
            response_data.append(entry)

        return {"results": response_data}

    @app_main.get("/metrics")
    def get_metrics():
        try:
            gpus = GPUtil.getGPUs()
            gpu_info = []
            for gpu in gpus:
                gpu_info.append({
                    "id": gpu.id,
                    "name": gpu.name,
                    "load": f"{gpu.load * 100:.2f}%",  # Format as percentage
                    "temperature": f"{gpu.temperature} C",
                    "memory": {
                        "used": f"{gpu.memoryUsed / 1024:.2f} GB",
                        "total": f"{gpu.memoryTotal / 1024:.2f} GB"
                    }
                })
            return {"gpu_info": gpu_info}
        except Exception as e:
            return {"error": str(e)}

def setup_worker_routes(): 
    @app_worker.get("/healthz")
    def health_check():
        if not torch.cuda.is_available():
            raise HTTPException(status_code=500, detail="No GPU available")
        if not generator:
            raise HTTPException(status_code=500, detail="Llama model not initialized")
        return {"status": "Healthy"}

    @app_worker.get("/metrics")
    def get_metrics():
        try:
            gpus = GPUtil.getGPUs()
            gpu_info = []
            for gpu in gpus:
                gpu_info.append({
                    "id": gpu.id,
                    "name": gpu.name,
                    "load": f"{gpu.load * 100:.2f}%",  # Format as percentage
                    "temperature": f"{gpu.temperature} C",
                    "memory": {
                        "used": f"{gpu.memoryUsed / 1024:.2f} GB",
                        "total": f"{gpu.memoryTotal / 1024:.2f} GB"
                    }
                })
            return {"gpu_info": gpu_info}
        except Exception as e:
            return {"error": str(e)}

def start_worker_server():
    print(f"Worker {dist.get_rank()} HTTP health server started at port 5000")
    uvicorn.run(app=app_worker, host='0.0.0.0', port=5000)

def worker_listen_tasks(): 
    while True:
        worker_num = dist.get_rank()
        print(f"Worker {worker_num} ready to recieve next command")
        config = [None] * 3  # Command and its associated data
        try: 
            dist.broadcast_object_list(config, src=0)
            command = config[0]

            if command == "text_generate":
                try:
                    input_string = config[1]
                    parameters = config[2]
                    generator.text_completion(
                        input_string,
                        max_gen_len=parameters.get('max_gen_len', None),
                        temperature=parameters.get('temperature', 0.6),
                        top_p=parameters.get('top_p', 0.9),
                    )
                except Exception as e:
                    print(f"Error in generation: {str(e)}")
            elif command == "shutdown":
                print(f"Worker {worker_num} shutting down")
                os.killpg(os.getpgrp(), signal.SIGTERM)
        except torch.distributed.DistBackendError as e:
            print("torch.distributed.DistBackendError", e)
            os.killpg(os.getpgrp(), signal.SIGTERM)
        except Exception as e:
            print(f"Error in Worker Listen Task", e)
            if 'Socket Timeout' in str(e):
                print("A socket timeout occurred.")
            os.killpg(os.getpgrp(), signal.SIGTERM)

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
        uvicorn.run(app=app_main, host='0.0.0.0', port=5000)  # Use the app_main instance.
    else:
        # This code is executed by all processes that aren't the globally ranked 0.
        # This includes processes on the main node as well as on other nodes.

        # Uncomment to enable worker logs
        # sys.stdout = sys.__stdout__

        os.setpgrp()
        try: 
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
        finally:
            # Additional fail-safe (to ensure no lingering processes)
            os.killpg(os.getpgrp(), signal.SIGTERM)