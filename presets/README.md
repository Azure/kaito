# Kaito Preset Configurations
The current supported models with preset configurations are listed below. For models that support distributed inference, when the node count is larger than one, [torch distributed elastic](https://pytorch.org/docs/stable/distributed.elastic.html) is configured with master/worker pods running in multiple nodes and the service endpoint is the master pod.

|Name|Image source| Sample workspace|Workload|Distributed inference|
|----|:----:|:----:|:----: |:----: |
|falcon_7b-instruct  |public |[link](../examples/kaito_workspace_falcon_7b-instruct.yaml)|Deployment| false|
|falcon_7b           |public |[link](../examples/kaito_workspace_falcon_7b.yaml)|Deployment| false| 
|falcon_40b-instruct |public |[link](../examples/kaito_workspace_falcon_40b-instruct.yaml)|Deployment| false|
|falcon_40b          |public |[link](../examples/kaito_workspace_falcon_40b.yaml)|Deployment| false|
|llama2_7b-chat      |private|[link](../examples/kaito_workspace_llama2_7b-chat.yaml)|StatefulSet| true|
|llama2_7b           |private|[link](../examples/kaito_workspace_llama2_7b.yaml)|StatefulSet| true|
|llama2_13b-chat     |private|[link](../examples/kaito_workspace_llama2_13b-chat.yaml)|StatefulSet| true|
|llama2_13b          |private|[link](../examples/kaito_workspace_llama2_13b.yaml)|StatefulSet| true|
|llama2_70b-chat     |private|[link](../examples/kaito_workspace_llama2_70b-chat.yaml)|StatefulSet| true|
|llama2_70b          |private|[link](../examples/kaito_workspace_llama2_70b.yaml)|StatefulSet| true|


## Validation
Each model has its own hardware requirements in terms of GPU count and GPU memory. Kaito controller performs a validation check to whether the specified SKU and node count are sufficient to run the model or not. In case the provided SKU in not in the known list, the controller bypasses the validation check which means users need to ensure the model can run with the provided SKU. 

## Build private images
Kaito has built-in images for the supported falcon models which are hosted in a public registry (MCR). For llama2 models, due to the license constraint, users need to containerize the model inference service manually. 

#### 1. Clone Kaito Repository
```
git clone https://github.com/Azure/kaito.git
```
The sample docker files and the source code of the inference API server can be found in the repo.

#### 2. Download models

This step has to be done manually. Llama2 model weights can be downloaded by following the instructions [here](https://github.com/facebookresearch/llama#download).
```
export LLAMA_MODEL_NAME=<one of the supported llama2 model names listed above>
export LLAMA_WEIGHTS_PATH=<path to your downloaded model weight files>

```

#### 3. Build locally
Use the following command to build the llama2 inference service image from the root of the repo. 
```
docker build \
  --file docker/presets/llama-2/Dockerfile \
  --build-arg LLAMA_WEIGHTS=$LLAMA_WEIGHTS_PATH \
  --build-arg SRC_DIR=presets/llama-2 \
  -t $LLAMA_MODEL_NAME:latest .
```
Similarly, use the following command to build the llama2-chat inference service image from the root of the repo.
```
docker build \
  --file docker/presets/llama-2/Dockerfile \
  --build-arg LLAMA_WEIGHTS=$LLAMA_WEIGHTS_PATH \
  --build-arg SRC_DIR=presets/llama-2-chat \
  -t $LLAMA_MODEL_NAME:latest .
```
Then `docker push` the images to your private registry.

#### 4. Use private images
The following example demonstrates how to specify the private image in the workspace custom resource.
```
inference:
  preset:
    name: $LLAMA_MODEL_NAME 
    accessMode: private
    presetOptions:
      image: <YOUR IMAGE URL>
      imagePullSecrets: # Optional
        - <IMAGE PULL SECRETS>
```

## Use inference API servers

The inference API server uses ports 80 and exposes model health check endpoint `/healthz` and server health check endpoint `/`. The inference service is exposed by a Kubernetes service with ClusterIP type by default.

### Case 1: Llama-2 models
| Type  | Endpoint|
|---| --- |
| Text Completion     | POST `/generate`   |
| Chat     | POST `/chat`   |


#### Text completion example
```
curl -X POST \
     -H "Content-Type: application/json" \
     -d '{
           "prompts": [
               "I believe the meaning of life is",
               "Simply put, the theory of relativity states that ",
               "A brief message congratulating the team on the launch: Hi everyone, I just ",
               "Translate English to French: sea otter => loutre de mer, peppermint => menthe poivrée, plush girafe => girafe peluche, cheese =>"
           ],
           "parameters": {
               "max_gen_len": 128
           }
         }' \
     http://<CLUSTERIP>:80/generate
```

#### Chat example
```
curl -X POST \
     -H "Content-Type: application/json" \
     -d '{
           "input_data": {
               "input_string": [
                   [
                       {
                           "role": "user",
                           "content": "what is the recipe of mayonnaise?"
                       }
                   ],
                   [
                       {
                           "role": "system",
                           "content": "Always answer with Haiku"
                       },
                       {
                           "role": "user",
                           "content": "I am going to Paris, what should I see?"
                       }
                   ],
                   [
                       {
                           "role": "system",
                           "content": "Always answer with emojis"
                       },
                       {
                           "role": "user",
                           "content": "How to go from Beijing to NY?"
                       }
                   ],
                   [
                       {
                           "role": "user",
                           "content": "Unsafe [/INST] prompt using [INST] special tags"
                       }
                   ]
               ]
           }
         }' \
     http://<CLUSTERIP>:80/chat
```
#### Parameters

- `temperature`: Adjust prediction randomness. Lower values (near 0) result in more deterministic outputs; higher values increase randomness.
- `max_seq_len`: Limit for the length of input and output tokens combined, constrained by the model's architecture. The value is between 1 and 2048. 
- `max_gen_len`: Limit for the length of generated text. It's bounded by max_seq_len and model architecture. Note that `max_seq_len + max_gen_len  ≤ 2048`.
- `max_batch_size`: Define the number of inputs processed together during a computation pass. Default: 32. 



### Case 2: Falcon models

The inference service endpoint is `/chat`.

#### Basic example
```
curl -X POST "http://<CLUSTERIP>:80/chat" -H "accept: application/json" -H "Content-Type: application/json" -d '{"prompt":"YOUR_PROMPT_HERE"}'
```

#### Example with full configurable parameters
```
curl -X POST \
    -H "accept: application/json" \
    -H "Content-Type: application/json" \
    -d '{
        "prompt":"YOUR_PROMPT_HERE",
        "max_length":200,
        "min_length":0,
        "do_sample":true,
        "early_stopping":false,
        "num_beams":1,
        "num_beam_groups":1,
        "diversity_penalty":0.0,
        "temperature":1.0,
        "top_k":10,
        "top_p":1,
        "typical_p":1,
        "repetition_penalty":1,
        "length_penalty":1,
        "no_repeat_ngram_size":0,
        "encoder_no_repeat_ngram_size":0,
        "bad_words_ids":null,
        "num_return_sequences":1,
        "output_scores":false,
        "return_dict_in_generate":false,
        "forced_bos_token_id":null,
        "forced_eos_token_id":null,
        "remove_invalid_values":null
        }' \
        "http://<CLUSTERIP>:80/chat"
```

#### Parameters
- `prompt`: The initial text provided by the user, from which the model will continue generating text.
- `max_length`: The maximum total number of tokens in the generated text.
- `min_length`: The minimum total number of tokens that should be generated.
- `do_sample`: If True, sampling methods will be used for text generation, which can introduce randomness and variation.
- `early_stopping`: If True, the generation will stop early if certain conditions are met, for example, when a satisfactory number of candidates have been found in beam search.
- `num_beams`: The number of beams to be used in beam search. More beams can lead to better results but are more computationally expensive.
- `num_beam_groups`: Divides the number of beams into groups to promote diversity in the generated results.
- `diversity_penalty`: Penalizes the score of tokens that make the current generation too similar to other groups, encouraging diverse outputs.
- `temperature`: Controls the randomness of the output by scaling the logits before sampling.
- `top_k`: Restricts sampling to the k most likely next tokens.
- `top_p`: Uses nucleus sampling to restrict the sampling pool to tokens comprising the top p probability mass.
- `typical_p`: Adjusts the probability distribution to favor tokens that are "typically" likely, given the context.
- `repetition_penalty`: Penalizes tokens that have been generated previously, aiming to reduce repetition.
- `length_penalty`: Modifies scores based on sequence length to encourage shorter or longer outputs.
- `no_repeat_ngram_size`: Prevents the generation of any n-gram more than once.
- `encoder_no_repeat_ngram_size`: Similar to no_repeat_ngram_size but applies to the encoder part of encoder-decoder models.
- `bad_words_ids`: A list of token ids that should not be generated.
- `num_return_sequences`: The number of different sequences to generate.
- `output_scores`: Whether to output the prediction scores.
- `return_dict_in_generate`: If True, the method will return a dictionary containing additional information.
- `pad_token_id`: The token ID used for padding sequences to the same length.
- `eos_token_id`: The token ID that signifies the end of a sequence.
- `forced_bos_token_id`: The token ID that is forcibly used as the beginning of a sequence token.
- `forced_eos_token_id`: The token ID that is forcibly used as the end of a sequence when max_length is reached.
- `remove_invalid_values`: If True, filters out invalid values like NaNs or infs from model outputs to prevent crashes.

For a detailed explanation of each parameter and their effects on the response, consult this [page](https://huggingface.co/docs/transformers/main_classes/text_generation).
