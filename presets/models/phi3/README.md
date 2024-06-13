## Supported Models
| Model name               | Model source | Sample workspace|Kubernetes Workload|Distributed inference|
|--------------------------|:----:|:----:| :----: |:----: |
| phi-3-mini-4k-instruct   |[microsoft](https://huggingface.co/microsoft/Phi-3-mini-4k-instruct)|[link](../../../examples/inference/kaito_workspace_phi-2.yaml)|Deployment| false|
| phi-3-mini-128k-instruct |[microsoft](https://huggingface.co/microsoft/Phi-3-mini-128k-instruct)|[link](../../../examples/inference/kaito_workspace_phi-2.yaml)|Deployment| false|


## Image Source
- **Public**: Kaito maintainers manage the lifecycle of the inference service images that contain model weights. The images are available in Microsoft Container Registry (MCR).

## Usage

Phi-3 Mini models are best suited for prompts using the chat format as follows. You can provide the prompt as a question with a generic template as follow:

```
<|user|>\nQuestion<|end|>\n<|assistant|>
```

For example:
```
<|user|>
How to explain Internet for a medieval knight?<|end|>
<|assistant|>

```

Or in the case of few shots prompt:
```
<|user|>
I am going to Paris, what should I see?<|end|>
<|assistant|>
Paris, the capital of France, is known for its stunning architecture, art museums, historical landmarks, and romantic atmosphere. Here are some of the top attractions to see in Paris:\n\n1. The Eiffel Tower: The iconic Eiffel Tower is one of the most recognizable landmarks in the world and offers breathtaking views of the city.\n2. The Louvre Museum: The Louvre is one of the world's largest and most famous museums, housing an impressive collection of art and artifacts, including the Mona Lisa.\n3. Notre-Dame Cathedral: This beautiful cathedral is one of the most famous landmarks in Paris and is known for its Gothic architecture and stunning stained glass windows.\n\nThese are just a few of the many attractions that Paris has to offer. With so much to see and do, it's no wonder that Paris is one of the most popular tourist destinations in the world."<|end|>
<|user|>
What is so great about #1?<|end|>
<|assistant|>
```

The inference service endpoint is `/chat`.


### Basic example
```
curl -X POST "http://<SERVICE>:80/chat" -H "accept: application/json" -H "Content-Type: application/json" -d '{"prompt":"YOUR_PROMPT_HERE"}'
```
```
curl -X POST "http://<SERVICE>:80/chat" -H "accept: application/json" -H "Content-Type: application/json" -d '{"prompt":"<|user|> How to explain Internet for a medieval knight?<|end|><|assistant|>"}'
```


### Example with full configurable parameters
```
curl -X POST \
    -H "accept: application/json" \
    -H "Content-Type: application/json" \
    -d '{
        "prompt":"YOUR_PROMPT_HERE",
        "return_full_text": false,
        "clean_up_tokenization_spaces": false, 
        "prefix": null,
        "handle_long_generation": null,
        "generate_kwargs": {
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
            }
        }' \
        "http://<SERVICE>:80/chat"
```

### Parameters
- `prompt`: The initial text provided by the user, from which the model will continue generating text.
- `return_full_text`: If False only generated text is returned, else full text is returned.
- `clean_up_tokenization_spaces`: True/False, determines whether to remove potential extra spaces in the text output.
- `prefix`: Prefix added to the prompt.
- `handle_long_generation`: Provides strategies to address generations beyond the model's maximum length capacity.
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
- `encoder_no_repeat_ngram_size`: Similar to `no_repeat_ngram_size` but applies to the encoder part of encoder-decoder models.
- `bad_words_ids`: A list of token ids that should not be generated.
- `num_return_sequences`: The number of different sequences to generate.
- `output_scores`: Whether to output the prediction scores.
- `return_dict_in_generate`: If True, the method will return a dictionary containing additional information.
- `pad_token_id`: The token ID used for padding sequences to the same length.
- `eos_token_id`: The token ID that signifies the end of a sequence.
- `forced_bos_token_id`: The token ID that is forcibly used as the beginning of a sequence token.
- `forced_eos_token_id`: The token ID that is forcibly used as the end of a sequence when max_length is reached.
- `remove_invalid_values`: If True, filters out invalid values like NaNs or infs from model outputs to prevent crashes.
