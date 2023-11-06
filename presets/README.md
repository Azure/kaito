# Containerize LLM models
This repository provides support to containerize open-source large language models (LLMs) such as Llama and Falcon. It includes a Python webserver that integrates the model libraries to offer a simple inference service. Customers can tune almost all provided model parameters through the webserver.

## Table of Contents
1. Prerequisites
2. Building Llama Model Images
3. API Documentation
   - Llama Model APIs
   - Falcon Model APIs
4. Model Parameters
   - Llama Model Parameters
   - Falcon Model Parameters
5. Conclusion

## Prerequisites
Each model has its own infrastructure requirements. Kaito controller performs a validation check to ensure your machine(s) has the necessary resources to run the model. For more information see [sku_configs](https://github.com/Azure/kaito/blob/main/api/v1alpha1/sku_config.go)


### Building the Llama image
1. Select Model Version: Identify the Llama model version to build, such as llama-2-7b or llama-2-7b-chat. Available models include `llama-2-7b, llama-2-13b, llama-2-70b, llama-2-7b-chat, llama-2-13b-chat and llama-2-70b-chat`.

2. Local Preset Path: Point to the local path of the model presets, which are found at the [kaito/presets/llama-2](https://github.com/Azure/kaito/tree/main/presets/llama-2) or [kaito/presets/llama-2-chat](https://github.com/Azure/kaito/tree/main/presets/llama-2-chat) directories for text and chat models, respectively.

3. Model Weights: Ensure your model weights are organized as /llama/<MODEL-VERSION> for the build process to include them in the Docker image.

4. Build Command:
Execute the Docker build command, replacing placeholders with actual values:

```
docker build \
  --build-arg LLAMA_VERSION=<MODEL-VERSION> \
  --build-arg SRC_DIR=<PATH-TO-LLAMA-PRESET> \
  -t <YOUR-IMAGE-NAME>:<YOUR-TAG> .
```
For example, to build the llama-2-7b model, the command would look like this:
```
docker build \
  --build-arg LLAMA_VERSION=llama-2-7b \
  --build-arg SRC_DIR=/home/kaito/presets/llama-2 \
  -t llama-2-7b:latest .
```

5. Check Image:
Confirm the image creation with `docker images`.

6. Deploy Image with Kaito: With the private image ready, integrate it into the Kaito Controller by updating the inferenceSpec in the deployment YAML file:
inference:
  preset:
    name: <MODEL-VERSION>
    accessMode: private
    presetOptions:
      image: <YOUR IMAGE URL>
      imagePullSecrets: # Optional
        - <IMAGE SECRETS>
Replace `<MODEL-VERSION>`, `<YOUR-IMAGE-URL>`, and `<IMAGE-PULL-SECRET>` with your specific details. For a reference implementation, see the example at [kaito_workspace_llama2_7b-chat.yaml](https://github.com/Azure/kaito/blob/main/examples/kaito_workspace_llama2_7b-chat.yaml)


## API Documentation

### Llama-2 Text Completion 
1. Server Health Check <br>
Endpoint: ```/``` <br>
Method: GET <br>
Purpose: Check if the server is running. <br>
Example: ```curl http://localhost:5000/```

2. Model Health Check <br>
Endpoint: ```/healthz``` <br>
Method: GET <br>
Purpose: Check if the model and GPU are properly initialized. <br>
Example: ```curl http://localhost:5000/healthz```

3. Shutdown <br>
Endpoint: ```/shutdown``` <br>
Method: POST <br>
Purpose: Shutdown server and program processes.  <br>
Example: ```curl -X POST http://localhost:5000/shutdown```

4. Complete Text <br>
Endpoint: ```/generate``` <br>
Method: POST <br>
Purpose: Complete text based on a given prompt. <br>
Example: 
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
     http://localhost:5000/generate
```

### Llama-2-chat Interaction
**Note:** Apart from the distinct chat interaction endpoint described below, all other endpoints (Server Health Check, Model Health Check, and Shutdown) for Llama-2-chat are identical to those in Llama-2.

Chat Interaction <br>
Endpoint: ```/chat``` <br>
Method: POST <br>
Purpose: Facilitates chat-based text interactions. <br>
Example:
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
     http://localhost:5000/chat
```
```
curl -X POST \
     -H "Content-Type: application/json" \
     -d '{
           "input_data": {
               "input_string": [
                   [
                       {
                           "role": "user",
                           "content": "I am going to Paris, what should I see?"
                       },
                       {
                           "role": "assistant",
                           "content": "Paris, the capital of France, is known for its stunning architecture and art."
                       },
                       {
                           "role": "user",
                           "content": "What is so great about its art?"
                       }
                   ]
               ],
               "parameters": {
                   "temperature": 0.6,
                   "top_p": 0.9
               }
           }
         }' \
     http://localhost:5000/chat
```
```
curl -X POST \
     -H "Content-Type: application/json" \
     -d '{
           "input_data": {
               "input_string": [
                   [
                       {
                           "role": "system",
                           "content": "You are a helpful, respectful and honest assistant. Always answer as helpfully as possible, while being safe."
                       },
                       {
                           "role": "user",
                           "content": "Write a brief birthday message to John"
                       }
                   ]
               ]
           }
         }' \
     http://localhost:5000/chat
```
### Falcon
Chat Interaction <br>
Endpoint: ```/chat``` <br>
Method: POST <br>
Purpose: Facilitates chat-based text interactions. <br>

Basic Example - Replace `YOUR_PROMPT_HERE` with your actual prompt:
```
curl -X POST "http://localhost:5000/chat" -H "accept: application/json" -H "Content-Type: application/json" -d '{"prompt":"YOUR_PROMPT_HERE"}'
```

Advanced Example with Configurable Parameters:
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
        "http://localhost:5000/chat"
```

## Model Parameters

### LLama Model Parameters
 - `temperature`: Adjusts prediction randomness. Lower values (near 0) result in more deterministic outputs; higher values increase randomness.
 - `max_seq_len`: Sets the limit for the length of input and output tokens combined, constrained by the model's architecture. Range: [1, 2048]. [See code] (https://github.com/facebookresearch/llama/blob/llama_v2/llama/model.py#L31)
- `max_gen_len`: Limits the length of generated text. It's bound by max_seq_len and architecture limit. Total of max_seq_len + max_gen_len must be ≤ 2048 [See code] (https://github.com/facebookresearch/llama/blob/llama_v2/llama/generation.py#L164).
- `max_batch_size`: Defines the number of inputs processed together during a computation pass. Larger sizes can improve training speed but consume more memory. Default: 32. (https://github.com/facebookresearch/llama/blob/llama_v2/llama/model.py#L30)

### Falcon Model Parameters
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

For a detailed explanation of each parameter and their effects on the response, consult this [reference page](https://huggingface.co/docs/transformers/main_classes/text_generation)

## Conclusion
These APIs provide a streamlined approach to harness the capabilities of the Llama 2 and Falcon models for text generation and chat-oriented applications. Ensure the correct deployment and configuration for optimal utilization.



