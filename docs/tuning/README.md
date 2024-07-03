# Kaito Tuning Workspace API

This guide provides instructions on how to use the Kaito Tuning Workspace API for parameter-efficient fine-tuning (PEFT) of models. The API supports methods like LoRA and QLoRA and allows users to specify their own datasets and configuration settings.

## Getting Started

To use the Kaito Tuning Workspace API, you need to define a Workspace custom resource (CR). Below are examples of how to define the CR and its various components.

## Example Workspace Definitions
Here are three examples of using the API to define a workspace for tuning different models:

Example 1: Tuning [`phi-3-mini`](../../examples/fine-tuning/kaito_workspace_tuning_phi_3.yaml)

Example 2: Tuning `falcon-7b`
```yaml
apiVersion: kaito.sh/v1alpha1
kind: Workspace
metadata:
  name: workspace-tuning-falcon
resource:
  instanceType: "Standard_NC6s_v3"
  labelSelector:
    matchLabels:
      app: tuning-phi-3-falcon
tuning:
  preset:
    name: falcon-7b
  method: qlora
  input:
    image: ACR_REPO_HERE.azurecr.io/IMAGE_NAME_HERE:0.0.1
    imagePullSecrets: 
      - IMAGE_PULL_SECRETS_HERE
  output:
    image: ACR_REPO_HERE.azurecr.io/IMAGE_NAME_HERE:0.0.1  # Tuning Output
    imagePushSecret: aimodelsregistrysecret

```
Generic TuningSpec Structure: 
```yaml
tuning:
  preset:
    name: preset-model
  method: lora or qlora
  config: custom-configmap (optional)
  input: # Image or URL
    urls:
      - "https://example.com/dataset.parquet?download=true"
  # Future updates will include support for v1.Volume as an input type.
  output: # Image
    image: "youracr.azurecr.io/custom-adapter:0.0.1"
    imagePushSecret: youracrsecret
```

## Default ConfigMaps
The default configuration for different tuning methods can be specified 
using ConfigMaps. Below are examples for LoRA and QLoRA methods.

- [Default LoRA ConfigMap](../../charts/kaito/workspace/templates/lora-params.yaml)

- [QLoRA ConfigMap](../../charts/kaito/workspace/templates/qlora-params.yaml)

## Using Custom ConfigMaps
You can specify your own custom ConfigMap and include it in the `Config` 
field of the `TuningSpec`

For more information on configurable parameters, please refer to the respective documentation links provided in the default ConfigMap examples.

## Key Parameters in Kaito ConfigMap Structure
The following sections highlight important parameters for each configuration area in the Kaito ConfigMap. For a complete list of parameters, please refer to the provided Hugging Face documentation links.

ModelConfig [Full List](https://huggingface.co/docs/transformers/v4.40.2/en/model_doc/auto#transformers.AutoModelForCausalLM.from_pretrained)
- torch_dtype: Specifies the data type for PyTorch tensors, e.g., "bfloat16".
- local_files_only: Indicates whether to only use local files.
- device_map: Configures device mapping for the model, typically "auto".

QuantizationConfig [Full List](https://huggingface.co/docs/transformers/v4.40.2/en/main_classes/quantization#transformers.BitsAndBytesConfig)
- load_in_4bit: Enables loading the model with 4-bit precision.
- bnb_4bit_quant_type: Specifies the type of 4-bit quantization, e.g., "nf4".
- bnb_4bit_compute_dtype: Data type for computation, e.g., "bfloat16".
- bnb_4bit_use_double_quant: Enables double quantization.

LoraConfig [Full List](https://huggingface.co/docs/peft/v0.8.2/en/package_reference/lora#peft.LoraConfig)
- r: Rank of the low-rank matrices used in LoRA.
- lora_alpha: Scaling factor for LoRA.
- lora_dropout: Dropout rate for LoRA layers.

TrainingArguments [Full List](https://huggingface.co/docs/transformers/v4.40.2/en/main_classes/trainer#transformers.TrainingArguments)
- output_dir: Directory where the training results will be saved.
- ddp_find_unused_parameters: Flag to handle unused parameters during distributed training.
- save_strategy: Strategy for saving checkpoints, e.g., "epoch".
- per_device_train_batch_size: Batch size per device during training.
- num_train_epochs: Total number of training epochs to perform, defaults to 3.0.

DataCollator [Full List](https://huggingface.co/docs/transformers/v4.40.2/en/main_classes/data_collator#transformers.DataCollatorForLanguageModeling)
- mlm: Masked language modeling flag.

DatasetConfig [Full List](https://github.com/Azure/kaito/blob/main/presets/tuning/text-generation/cli.py#L44)
- shuffle_dataset: Whether to shuffle the dataset.
- train_test_split: Proportion of data used for training, typically set to 1 for using all data.

## Expected Dataset Format
The dataset should follow a specific format. For example:

- Conversational Format
  ```json
  {
    "messages": [
      {"role": "system", "content": "Marv is a factual chatbot that is also sarcastic."},
      {"role": "user", "content": "What's the capital of France?"},
      {"role": "assistant", "content": "Paris, as if everyone doesn't know that already."}
    ]
  }
  ```
For a practical example, refer to [HuggingFace Dolly 15k OAI-style dataset](https://huggingface.co/datasets/philschmid/dolly-15k-oai-style/tree/main).

- Instruction Format
  ```json
  {"prompt": "<prompt text>", "completion": "<ideal generated text>"}
  {"prompt": "<prompt text>", "completion": "<ideal generated text>"}
  {"prompt": "<prompt text>", "completion": "<ideal generated text>"}
  ```

For a practical example, refer to [HuggingFace Instruction Dataset](https://huggingface.co/datasets/HuggingFaceH4/instruction-dataset/tree/main)


## Dataset Format Support

The SFTTrainer supports popular dataset formats, allowing direct passage of the dataset without pre-processing. Supported formats include conversational and instruction formats. If your dataset is not in one of these formats, it will be passed directly to the SFTTrainer without any preprocessing. This may result in undefined behavior if the dataset does not align with the trainer's expected input structure.

To ensure proper function, you may need to preprocess the dataset to match one of the supported formats.

For example usage and more details, refer to the [Official Hugging Face documentation and tutorials](https://huggingface.co/docs/trl/v0.9.4/sft_trainer#dataset-format-support).