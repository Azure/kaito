# Copyright (c) Microsoft Corporation.
# Licensed under the MIT license.
import csv
import argparse
import requests
from datetime import datetime
import time
import uuid
import asyncio
import aiohttp
import sys
import time

def get_args():
    parser = argparse.ArgumentParser()
    parser.add_argument("--model", type=str, required=True)
    parser.add_argument("--num_nodes", type=int, required=True)
    parser.add_argument("--num_processes", type=int, required=True)
    parser.add_argument("--num_gpus", type=int, required=True)
    parser.add_argument("--num_prompts", type=int, required=True)
    parser.add_argument("--model_parallelism", type=str, required=True)
    parser.add_argument("--data_parallelism", type=str, required=True)
    parser.add_argument("--quantization", type=str, required=True)
    parser.add_argument("--machine", type=str, required=True)
    parser.add_argument("--use_accelerator", action='store_true', help="Use the Accelerator for parallel processing.")
    return parser.parse_args()

args = get_args()


async def async_request(session, request):
    try:
        start_time = time.time()
        timeout = aiohttp.ClientTimeout(total=600)
        print("MADE REQUEST")
        async with session.post('http://20.84.13.124:5000/chat', json={"prompt": request}, headers={'Content-Type': 'application/json'}, timeout=timeout) as response:
            response_text = await response.text()
            end_time = time.time()
        return response_text, end_time - start_time, len(request)
    except Exception as e:
        print(f"Request failed: {e}")
        return "Error", 0, len(request)

async def async_inference(prompts, writer, args, file):
    async with aiohttp.ClientSession() as session:
        tasks = [async_request(session, prompts[i]) for i in range(0, len(prompts))]
        time.sleep(1)
        responses = await asyncio.gather(*tasks, return_exceptions=True)
        # Execute tasks with a slight delay between each
        # responses = []
        # for task in tasks:
        #     response = await task
        #     responses.append(response)
            # await asyncio.sleep(0.1)  # Adds a 0.1-second delay between requests

    for response_text, inference_time, req_len in responses:
        timestamp = datetime.now().strftime('%Y-%m-%d %H:%M:%S')

        print(f"Response from the server: {response_text}")

        result = {
            "model": args.model,
            "num_nodes": args.num_nodes,
            "num_processes": args.num_processes,
            "num_gpus": args.num_gpus,
            "num_prompts": args.num_prompts,
            "prompt_len": req_len,
            "model_parallelism": args.model_parallelism,
            "data_parallelism": args.data_parallelism,
            "quantization": args.quantization,
            "machine": args.machine,
            "inference_time": inference_time,
            "request_id": str(uuid.uuid4()), # Generate a unique UUID
            "timestamp": timestamp
        }
        writer.writerow(result)
        file.flush()


def sync_inference(prompts, writer, args):
    for request in prompts:
        start_time = time.time()

        # Prepare the data for the POST request
        data = {"prompt": request}

        # Make the HTTP POST request
        response = requests.post('http://20.84.13.124:5000/chat', json=data, headers={'Content-Type': 'application/json'})

        end_time = time.time()
        inference_time = end_time - start_time
        timestamp = datetime.now().strftime('%Y-%m-%d %H:%M:%S')

        print(f"Response from the server: {response.text}")

        result = {
            "model": args.model,
            "num_nodes": args.num_nodes,
            "num_processes": args.num_processes,
            "num_gpus": args.num_gpus,
            "num_prompts": args.num_prompts,
            "prompt_len": len(request),
            "model_parallelism": args.model_parallelism,
            "data_parallelism": args.data_parallelism,
            "quantization": args.quantization,
            "machine": args.machine,
            "inference_time": inference_time,
            "request_id": str(uuid.uuid4()), # Generate a unique UUID
            "timestamp": timestamp
        }
        writer.writerow(result)

with open("../common-gpt-questions.csv", "r") as f:
    prompts = [line.strip() for line in f.readlines()]


def sync_main(): 
    fieldnames = ["model", "num_nodes", "num_processes", "num_gpus", "num_prompts", "prompt_len", "model_parallelism", "data_parallelism", "quantization", "machine", "inference_time", "request_id", "timestamp"]

    with open("/Users/ishaansehgal/Documents/kaito-ishaansehgal99/presets/test/falcon-benchmark/results/pytorch/sync-one-gpu-V100-ds.csv", "a", newline='') as f:
        writer = csv.DictWriter(f, fieldnames=fieldnames)
        writer.writeheader()
        sync_inference(prompts, writer, args)

async def async_main():
    fieldnames = ["model", "num_nodes", "num_processes", "num_gpus", "num_prompts", "prompt_len", "model_parallelism", "data_parallelism", "quantization", "machine", "inference_time", "request_id", "timestamp"]

    with open("/Users/ishaansehgal/Documents/kaito-ishaansehgal99/presets/test/falcon-benchmark/results/pytorch/async-one-gpu.csv", "a", newline='') as f:
        writer = csv.DictWriter(f, fieldnames=fieldnames)
        writer.writeheader()
        await async_inference(prompts, writer, args, f)

if __name__ == "__main__":
    sync_main()
    # asyncio.run(async_main())
