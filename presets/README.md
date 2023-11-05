# Containerize LLM models
This repository provides support to containerize open-source large language models (LLMs) such as Llama and Falcon. It includes a Python webserver that integrates the model libraries to offer a simple inference service. Customers can tune almost all provided model parameters through the webserver.

## Table of Contents
1. Prerequisites
2. Building Images
   - Building Llama Model Images
   - Building Falcon Model Images
3. Running the Models
   - Running Llama
   - Running Falcon
4. API Documentation
   - Llama Model APIs
   - Falcon Model APIs
5. Model Parameters
   - Llama Model Parameters
   - Falcon Model Parameters
6. Contributing
7. License


## Prerequisites
Each model has its own infrastructure requirements. Kaito controller performs a validation check to ensure your machine(s) has the necessary resources to run the model. For more information see [sku_configs](https://github.com/Azure/kaito/blob/main/api/v1alpha1/sku_config.go)


## Building Images 

1. Choose the Desired Model Directory: Navigate to either the llama-2 or llama-2-chat directory, based on the desired model.
2. Build the Docker Image: ```docker build -t your-image-name:your-tag .```
3. Deploy the Image to a Container: ```docker run --name your-container-name your-image-name:your-tag```


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

## Conclusion
These APIs provide a streamlined approach to harness the capabilities of the Llama 2 model for both text generation and chat-oriented applications. Ensure the correct deployment and configuration for optimal utilization.



