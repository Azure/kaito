#!/bin/bash

if [ -z "$1" ]; then
    echo "Usage: $0 <device_number>"
    exit 1
fi

device_number=$1

max_memory_usage=0

# Log file to store the memory usage, including the device number
log_file="gpu_memory_usage_${device_number}.log"

update_max_memory_usage() {
    current_memory_usage=$(nvidia-smi --query-gpu=memory.used --format=csv,noheader,nounits -i $device_number | tr -d '[:space:]')

    # Update the maximum memory usage if the current usage is higher
    if (( current_memory_usage > max_memory_usage )); then
        max_memory_usage=$current_memory_usage
        # Log the new maximum memory usage
        echo "$(date): New max GPU memory usage on GPU $device_number: ${max_memory_usage}MiB" | tee -a $log_file
    fi
}

# Monitor GPU memory usage every second
while true; do
    update_max_memory_usage
    sleep 1
done
