File is to explain the parameters used by llama-2

-> ckpt_dir: The directory containg the models checkpoint

-> tokenizer_path: path to the tokenizer model 

-> temperature: This hyperparam controls the randomness of predictions by scaling logits before applying the softmax for results. A temperature closer to 0 makes the model more deterministic (choosing the most probable output) whereas a higher temperature makes the output more random. 

-> max_seq_len: maximum length for both the input and output tokens. This is usually bound by the model's architecture (e.g. the maximum length the llama transformer can take at a time) so it is in range of [1, max_length_supported_by_model], (https://github.com/facebookresearch/llama/blob/main/llama/model.py#L31)

-> max_gen_len: A cap on maximum sequence length for the generated text. Like max_seq_len this is also usually bound by the model architecture. In this case it is in range of [1, 2048] (https://github.com/facebookresearch/llama/blob/main/llama/generation.py#L191) 

* Important to note that max_seq_len + max_gen_len must be <= to arch limit of 2048. So for example if you provide a 2048 input you will get 0 tokens as output. 

-> max_batch_size: maximum batch size for processing. This refers to the number of input example processed simaltaneously during a forward/backward pass. Larger batch sizes usually provide more accurate gradient estimate and generally lead to faster training but require more memory and can cause out-of-memory errors if set too high. In this case llama sets max_batch_size to 32 (https://github.com/facebookresearch/llama/blob/main/llama/model.py#L30)