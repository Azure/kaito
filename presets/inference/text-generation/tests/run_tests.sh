#!/bin/bash

echo "Running text-generation tests..."
python3 test_inference_api.py --pipeline text-generation --pretrained_model_name_or_path microsoft/phi-2 --allow_remote_files True

echo "Running conversational tests..."
python3 test_inference_api.py --pipeline conversational --pretrained_model_name_or_path mistralai/Mistral-7B-Instruct-v0.2 --allow_remote_files True

echo "Running invalid-pipeline tests..."
python3 test_inference_api.py --pipeline invalid-pipeline --pretrained_model_name_or_path mistralai/Mistral-7B-Instruct-v0.2 --allow_remote_files True
