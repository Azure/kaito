Exploring Memory Usage for Fine-Tuning LLMs

LLMs have millions, billions or even trillions of parameters. Hence training LLMs is a resource-intensive, and experimental process. For example:
- Google's Gemini Ultra: Reportedly cost $191M to train
- OpenAI's GPT-4: Estimated to cost >$100M to train, and 5-6 months
- DBRX from Databricks: Reportedly cost $10M to train

Asides from large corporations, rarely are individuals training LLMs from scratch. Rather Parameter Efficient Fine-Tuning (PeFT) methods were proposed to train LLMs on specified tasks efficiently. These methods are designed to greatly reduce the amount of resources (time, money, hardware) required to tune LLMs to perform better on specific tasks.

Today we will be walking through fine-tuning model Phi-3-mini-128K instruct on the [Databricks Dolly Dataset](https://huggingface.co/datasets/philschmid/dolly-15k-oai-style). We will be navigating the complex landscape of memory usage, and exploring tradeoffs on efficiently performing fine tuning. 

All experiments will be conducted using a NVIDA A100(80GB) GPU. 

## Phi-3 Model Architecture
- 32 layers (0 to 31) - Each layer contains Attention and MLP components
- Layer Structure (repeated for all 32 layers)
  ```
  model.layers.[i]:
      self_attn:
          o_proj.weight
          qkv_proj.weight
      mlp:
          gate_up_proj.weight
          down_proj.weight
      input_layernorm.weight
      post_attention_layernorm.weight
  ```
This serves as a great simple example for further exploration. Lets break down these layers further to try and really understand how this model is architected.

# Phi-3 Model Analysis

## Overall Model Statistics
- **Total Parameters:** 3,821,079,552
- **Trainable Parameters:** 3,821,079,552
- **Model Size:** 14,576.26 MB

## Model Structure Breakdown

### Component-wise Summary
- **Embedding:**
  - Parameters: 98,500,608 (2.58% of total)
  - Size: 375.75 MB (2.58% of total)
- **Attention:**
  - Parameters: 1,208,057,856 (31.62% of total)
  - Size: 4,608.38 MB (31.62% of total)
- **MLP:**
  - Parameters: 2,415,919,104 (63.23% of total)
  - Size: 9,216.00 MB (63.23% of total)
- **LayerNorm:**
  - Parameters: 101,376 (0.00% of total)
  - Size: 0.39 MB (0.00% of total)
- **Head:**
  - Parameters: 98,500,608 (2.58% of total)
  - Size: 375.75 MB (2.58% of total)

#### Layer-wise Summary
- **Layer Configuration:** 32 identical layers
  - Parameters per layer: 113,252,352 (2.96% of total)
  - Size per layer: 432.02 MB (2.96% of total)
- **Total for all layers:**
  - Parameters: 3,624,075,264 (94.84% of total)
  - Size: 13,824.75 MB (94.84% of total)

#### Non-Layer Parameters
- **model.embed_tokens.weight:**
  - Parameters: 98,500,608 (2.58% of total)
  - Size: 375.75 MB (2.58% of total)
- **model.norm.weight:**
  - Parameters: 3,072 (0.00% of total)
  - Size: 0.01 MB (0.00% of total)
- **lm_head.weight:**
  - Parameters: 98,500,608 (2.58% of total)
  - Size: 375.75 MB (2.58% of total)
- **Total Non-Layer Parameters:** 197,004,288 (5.16% of total)
- **Total Non-Layer Size:** 751.51 MB (5.16% of total)

For a further detailed breakdown on model parameters checkout [TODO](linktomodel_info.txt)

Conclusion: 

1. Model Architecture:
- The Phi-3 model consists of 32 identical layers, each containing about 113 million parameters.
- The repeating layer structure accounts for 94.84% of the model's total parameters and size.

2. Parameter Distribution:
- MLP components dominate the model, accounting for 63.23% of all parameters.
- Attention mechanisms are the second largest component, with 31.62% of parameters.
- Embedding and output head each contribute 2.58% of the total parameters.

3. Attention to Computation Ratio:
 - The ratio of attention (31.62%) to MLP (63.23%) parameters is approximately 1:2, which is common in many transformer architectures.




# Model Memory

We'll start by taking a look at [Accelerate's Memory Estimation](https://huggingface.co/docs/accelerate/en/usage_guides/model_size_estimator) for Loading `microsoft/Phi-3-mini-128k-instruct`
┌──────────────────────────────────────────────────────────────────┐
│  Memory Usage for loading `microsoft/Phi-3-mini-128k-instruct`   │
├───────┬─────────────┬──────────┬─────────────────────────────────┤
│ dtype │Largest Layer│Total Size│       Training using Adam       │
├───────┼─────────────┼──────────┼─────────────────────────────────┤
│float32│  432.02 MB  │ 14.23 GB │             56.94 GB            │
│float16│  216.01 MB  │ 7.12 GB  │             28.47 GB            │
│  int8 │  108.01 MB  │ 3.56 GB  │               N/A               │
│  int4 │   54.0 MB   │ 1.78 GB  │               N/A               │
└───────┴─────────────┴──────────┴─────────────────────────────────┘

Let's compare this with the actual memory used for loading the model

Float32 Precision - **15.73GB**
+-----------------------------------------------------------------------------------------+
| NVIDIA-SMI 550.90.07              Driver Version: 550.90.07      CUDA Version: 12.4     |
|-----------------------------------------+------------------------+----------------------+
| GPU  Name                 Persistence-M | Bus-Id          Disp.A | Volatile Uncorr. ECC |
| Fan  Temp   Perf          Pwr:Usage/Cap |           Memory-Usage | GPU-Util  Compute M. |
|                                         |                        |               MIG M. |
|=========================================+========================+======================|
|   0  NVIDIA A100 80GB PCIe          On  |   00000001:00:00.0 Off |                    0 |
| N/A   35C    P0             62W /  300W |   15005MiB /  81920MiB |      0%      Default |
|                                         |                        |             Disabled |
+-----------------------------------------+------------------------+----------------------+
- TODO: You'll note its slightly larger than model size (explain why use blog as example)

Float16 Precision - **8.09GB**
+-----------------------------------------------------------------------------------------+
| NVIDIA-SMI 550.90.07              Driver Version: 550.90.07      CUDA Version: 12.4     |
|-----------------------------------------+------------------------+----------------------+
| GPU  Name                 Persistence-M | Bus-Id          Disp.A | Volatile Uncorr. ECC |
| Fan  Temp   Perf          Pwr:Usage/Cap |           Memory-Usage | GPU-Util  Compute M. |
|                                         |                        |               MIG M. |
|=========================================+========================+======================|
|   0  NVIDIA A100 80GB PCIe          On  |   00000001:00:00.0 Off |                    0 |
| N/A   35C    P0             62W /  300W |    7717MiB /  81920MiB |      0%      Default |
|                                         |                        |             Disabled |
+-----------------------------------------+------------------------+----------------------+

BFloat16 Precision - **8.09GB**
+-----------------------------------------------------------------------------------------+
| NVIDIA-SMI 550.90.07              Driver Version: 550.90.07      CUDA Version: 12.4     |
|-----------------------------------------+------------------------+----------------------+
| GPU  Name                 Persistence-M | Bus-Id          Disp.A | Volatile Uncorr. ECC |
| Fan  Temp   Perf          Pwr:Usage/Cap |           Memory-Usage | GPU-Util  Compute M. |
|                                         |                        |               MIG M. |
|=========================================+========================+======================|
|   0  NVIDIA A100 80GB PCIe          On  |   00000001:00:00.0 Off |                    0 |
| N/A   35C    P0             62W /  300W |    7717MiB /  81920MiB |      0%      Default |
|                                         |                        |             Disabled |
+-----------------------------------------+------------------------+----------------------+

TODO: explain why these are the same


Enabling 8-bit Quantization (`load_in_8bit=True`) - **4.64GB**
+-----------------------------------------------------------------------------------------+
| NVIDIA-SMI 550.90.07              Driver Version: 550.90.07      CUDA Version: 12.4     |
|-----------------------------------------+------------------------+----------------------+
| GPU  Name                 Persistence-M | Bus-Id          Disp.A | Volatile Uncorr. ECC |
| Fan  Temp   Perf          Pwr:Usage/Cap |           Memory-Usage | GPU-Util  Compute M. |
|                                         |                        |               MIG M. |
|=========================================+========================+======================|
|   0  NVIDIA A100 80GB PCIe          On  |   00000001:00:00.0 Off |                    0 |
| N/A   33C    P0             62W /  300W |    4425MiB /  81920MiB |      0%      Default |
|                                         |                        |             Disabled |
+-----------------------------------------+------------------------+----------------------+


Enabling 4-bit Quantization (`load_in_4bit=True`) - **3.00 GB**
+-----------------------------------------------------------------------------------------+
| NVIDIA-SMI 550.90.07              Driver Version: 550.90.07      CUDA Version: 12.4     |
|-----------------------------------------+------------------------+----------------------+
| GPU  Name                 Persistence-M | Bus-Id          Disp.A | Volatile Uncorr. ECC |
| Fan  Temp   Perf          Pwr:Usage/Cap |           Memory-Usage | GPU-Util  Compute M. |
|                                         |                        |               MIG M. |
|=========================================+========================+======================|
|   0  NVIDIA A100 80GB PCIe          On  |   00000001:00:00.0 Off |                    0 |
| N/A   35C    P0             62W /  300W |    2861MiB /  81920MiB |      0%      Default |
|                                         |                        |             Disabled |
+-----------------------------------------+------------------------+----------------------+
 

Conclusion: 

┌───────────────────────────────────────────┐
│                Memory Usage               │
├───────┬─────────────┬──────────-┬─────────┤
│ dtype │Largest Layer│ Predicted |  Actual │
├───────┼─────────────┼──────────-┼───────-─┤
│float32│  432.02 MB  │ 14.23 GB  │ 15.73GB │
│float16│  216.01 MB  │ 7.12 GB   │  8.09GB │
│  int8 │  108.01 MB  │ 3.56 GB   │ 4.64GB* │
│  int4 │   54.0 MB   │ 1.78 GB   │ 3.00GB* │
└───────┴─────────────┴──────────-┴─────────┘


*In reality enabling 4-Bit/8-Bit Quantization results in Mixed Precision. As follows:
```
model.layers.[i]:
    self_attn:
        o_proj.weight: INT4/INT8
        qkv_proj.weight: INT4/INT8
    mlp:
        gate_up_proj.weight: INT4/INT8
        down_proj.weight: INT4/INT8
    input_layernorm.weight: Float16
    post_attention_layernorm.weight: Float16
```

This is beneficial for 
1. **Memory Efficiency**: Quantization for large matrix operations (Attention & MLP) significantly reduces memory footprint.
2. **Precision Balance**: Critical components (embeddings, normalization, output) retain higher precision to maintain accuracy.
3. **Consistency**: The quantization pattern is uniform across all layers, simplifying implementation and analysis.





When training on a batch size of 1, each stage of the training process is expected to have near the following memory results for each precision:
┌─────────────────────────────────────────────────────────────────────────────┐
│       Training using Adam for 'microsoft/Phi-3-mini-128k-instruct'          │
├────────────────┬────────────┬────────────────────┬───────────────┬──────────┤
│     dtype      │   Model    │  Grad Calculation  │ Backward pass │Optim step│
├────────────────┼────────────┼────────────────────┼───────────────┼──────────┤
│    float32     │  14.23 GB  │      14.23 GB      │    28.47 GB   │ 56.94 GB │
│float16/bfloat16│  14.23 GB  │      21.35 GB      │    28.47 GB   │ 28.47 GB │
└────────────────┴────────────┴────────────────────┴───────────────┴──────────┘



These metrics will serve as our benchmark to achieve.


We will start by tuning with LoRA and tweaking hyperparameters (batch size, optimizers, rank values, and target_modules) to see how these affect memory usage.








