# import requests
# import asyncio
# import time
# # import pynvml
# from concurrent.futures import ThreadPoolExecutor
#
# # Initialize NVML for GPU monitoring
# # pynvml.nvmlInit()
#
# # Function to get GPU metrics
# # def get_gpu_metrics():
# #     device_count = pynvml.nvmlDeviceGetCount()
# #     metrics = []
# #     for i in range(device_count):
# #         handle = pynvml.nvmlDeviceGetHandleByIndex(i)
# #         util = pynvml.nvmlDeviceGetUtilizationRates(handle)
# #         memory = pynvml.nvmlDeviceGetMemoryInfo(handle)
# #         metrics.append({
# #             f"GPU_{i}": {
# #                 "Utilization": f"{util.gpu}%",
# #                 "Memory_Total_MiB": memory.total / (1024 ** 2),
# #                 "Memory_Used_MiB": memory.used / (1024 ** 2),
# #                 "Memory_Free_MiB": memory.free / (1024 ** 2)
# #             }
# #         })
# #     return metrics
#
# # Endpoint URL
# SERVICE_IP = "52.226.168.2"
# ENDPOINT_URL = f"http://{SERVICE_IP}:80/chat"
#
# # Sample request payload
# PAYLOAD = {
#     "prompt": "RANDOM PROMPT",
#     "return_full_text": False,
#     "clean_up_tokenization_spaces": False,
#     "prefix": None,
#     "handle_long_generation": None,
#     "generate_kwargs": {"max_length": 200}
# }
#
# # Headers
# HEADERS = {
#     "accept": "application/json",
#     "Content-Type": "application/json"
# }
#
# # Function to send a single request and log GPU metrics
# def send_request_and_log_gpu_metrics():
#     # pre_metrics = get_gpu_metrics()  # Get GPU metrics before the request
#     response = requests.get(f"http://{SERVICE_IP}:80/")
#     # response = requests.post(ENDPOINT_URL, json=PAYLOAD, headers=HEADERS)
#     # post_metrics = get_gpu_metrics()  # Get GPU metrics after the request
#     response_time = response.elapsed.total_seconds()
#     return response_time
#
# # Asynchronous function to send requests concurrently and log GPU metrics
# async def send_requests_concurrently(num_requests: int):
#     loop = asyncio.get_event_loop()
#     with ThreadPoolExecutor() as executor:
#         tasks = [loop.run_in_executor(executor, send_request_and_log_gpu_metrics) for _ in range(num_requests)]
#         completed, pending = await asyncio.wait(tasks)
#         times = [t.result() for t in completed]
#         return times
#
# # Main function to orchestrate the test
# async def main():
#     num_requests = 100000000000  # Adjustable
#     print(num_requests)
#     start_time = time.time()
#     response_times = await send_requests_concurrently(num_requests)
#     end_time = time.time()
#
#     print(f"Total time for {num_requests} requests: {end_time - start_time} seconds")
#     print(f"Average response time: {sum(response_times) / len(response_times)} seconds")
#     print(f"Max response time: {max(response_times)} seconds")
#     print(f"Min response time: {min(response_times)} seconds")
#
# # Run the test
# if __name__ == "__main__":
#     asyncio.run(main())

import subprocess
from multiprocessing import Pool

# Endpoint URL
SERVICE_IP = "52.226.168.2"
ENDPOINT_URL = f"http://{SERVICE_IP}:80/"
CHAT_ENDPOINT_URL = f"http://{SERVICE_IP}:80/chat"

num_requests = 20  # Adjust based on your testing needs

# Function to be executed in each separate process
def make_request(_):
    try:
        print("Making Request")
        result = subprocess.run(["curl", "-s", "-o", "/dev/null", "-w", "%{time_total}", ENDPOINT_URL], capture_output=True, text=True)
        response_time = result.stdout.strip()
        return float(response_time)
    except Exception as e:
        print(f"Error making request: {e}")
        return None

def make_chat_request(_):
    try:
        print("Making Chat Request")
        json_payload = '{"prompt":"RANDOM PROMPT","return_full_text":false,"clean_up_tokenization_spaces":false,"prefix":null,"handle_long_generation":null,"generate_kwargs":{"max_length":200}}'

        result = subprocess.run([
            "curl", "-s", "-X", "POST",
            "-H", "accept: application/json",
            "-H", "Content-Type: application/json",
            "-d", json_payload,
            "-o", "/dev/null", "-w", "%{time_total}",
            CHAT_ENDPOINT_URL
        ], capture_output=True, text=True)

        response_time = result.stdout.strip()
        return float(response_time)
    except Exception as e:
        print(f"Error making chat request: {e}")
        return None


def main():
    with Pool(processes=num_requests) as pool:
        # Running them in parallel
        response_times = pool.map(make_chat_request, range(num_requests))

        response_times = [time for time in response_times if time is not None]

        if response_times:
            print(f"Average response time: {sum(response_times) / len(response_times)} seconds")
            print(f"Max response time: {max(response_times)} seconds")
            print(f"Min response time: {min(response_times)} seconds")
        else:
            print("No valid response times were collected.")

# Run the test
if __name__ == "__main__":
    main()
