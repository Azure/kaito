#!/bin/bash

# Initialize the maximum memory usage variable
max_memory_usage=0

# Log file to store the memory usage
log_file="gpu_memory_usage.log"

# Function to update the maximum memory usage
update_max_memory_usage() {
    # Extract the current memory usage from nvidia-smi and trim any spaces or newlines
    current_memory_usage=$(nvidia-smi --query-gpu=memory.used --format=csv,noheader,nounits | tr -d '[:space:]')

    # Update the maximum memory usage if the current usage is higher
    if (( current_memory_usage > max_memory_usage )); then
        max_memory_usage=$current_memory_usage
        # Log the new maximum memory usage
        echo "$(date): New max GPU memory usage: ${max_memory_usage}MiB" | tee -a $log_file
    fi
}

# Monitor GPU memory usage every second
while true; do
    update_max_memory_usage
    sleep 1
done
