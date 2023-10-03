# Falcon Inference Benchmarking

Benchmark and analyze the performance of Falcon's model inference with our utilities.

## 📂 Contents

- **benchmark-inference.py**: Runs the inference benchmarking and outputs a CSV containing performance metrics.
- **plot.ipynb**: A Jupyter Notebook that visualizes the results of the benchmarking from the CSV.

## 🚀 Getting Started

### Prerequisites

Ensure your `accelerate` configuration aligns with the values provided during benchmarking. To check or set up your configuration, you can run the command: `accelerate config`.

### Running the Benchmark

To run the benchmarking, use the `accelerate launch` command along with the provided script and parameters. For example:

```bash
accelerate launch inference-benchmark.py --model MODEL_NAME --num_nodes NODE_COUNT --num_processes PROCESS_COUNT --num_gpus GPU_COUNT --num_prompts PROMPT_COUNT --model_parallelism PARALLELISM_TYPE --data_parallelism DATA_PARALLELISM_TYPE --quantization QUANTIZATION_TYPE --machine MACHINE_TYPE
```

For example: 
```bash
accelerate launch inference-benchmark.py --model falcon-40b --num_nodes 1 --num_processes 1 --num_gpus 1 --num_prompts 1 --model_parallelism deepspeed --data_parallelism none --quantization bf16 --machine Standard_NC96ads_A100_v4
```

### 📊 Visualizing the Results
Open the `plot.ipynb` notebook in Jupyter. Ensure you have the necessary libraries installed. The notebook will guide you through loading the benchmark data and visualizing the results.

Note: The notebook supports the visualization of results from multiple experiments, as shown in the provided snippets. Make sure to update the data source if you have benchmark results from other experiments or configurations.

### 📊 Output
- The benchmarking script (inference-benchmark.py) will produce a results.csv file containing the performance statistics of each inference request.
- The visualization notebook (plot.ipynb) provides plots that offer insights into the inference performance across various configurations and parameters.

### 📌 Notes
- Ensure that you have a file named requests.csv in the directory before running the benchmark. This file should contain the prompts for which you'd like to benchmark inference.
- The script and notebook are set up to work with models hosted on HuggingFace, specifically "tiiuae/falcon-7b-instruct" and "tiiuae/falcon-40b-instruct". Modify the model_id in the script if you want to benchmark a different model.

