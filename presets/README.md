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
5. Contributing
6. License
7. Conclusion

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
        "no_repeat_ngram_size":0,"encoder_no_repeat_ngram_size":0,"bad_words_ids":null,
        "num_return_sequences":1,
        "output_scores":false,"return_dict_in_generate":false,"forced_bos_token_id":null,"forced_eos_token_id":null,"remove_invalid_values":null
        }' \
        "http://localhost:5000/chat"
```


## Conclusion
These APIs provide a streamlined approach to harness the capabilities of the Llama 2 model for both text generation and chat-oriented applications. Ensure the correct deployment and configuration for optimal utilization.



