# Copyright (c) Microsoft Corporation.
# Licensed under the MIT license.
import os
from dataclasses import asdict, dataclass, field
from typing import Annotated, Any, Dict, List, Optional

import GPUtil
import psutil
import torch
import transformers
import subprocess
import uvicorn
from fastapi import Body, FastAPI, HTTPException
from fastapi.responses import Response
from pydantic import BaseModel, Extra, Field, validator
from transformers import (AutoModelForCausalLM, AutoTokenizer,
                          GenerationConfig, HfArgumentParser)
from peft import PeftModel

@dataclass
class ModelConfig:
    """
    Transformers Model Configuration Parameters
    """
    pipeline: str = field(metadata={"help": "The model pipeline for the pre-trained model"})
    pretrained_model_name_or_path: Optional[str] = field(default="/workspace/tfs/weights", metadata={"help": "Path to the pretrained model or model identifier from huggingface.co/models"})
    combination_type: Optional[str]=field(default="svd", metadata={"help": "The combination type of multi adapters"})
    state_dict: Optional[Dict[str, Any]] = field(default=None, metadata={"help": "State dictionary for the model"})
    cache_dir: Optional[str] = field(default=None, metadata={"help": "Cache directory for the model"})
    from_tf: bool = field(default=False, metadata={"help": "Load model from a TensorFlow checkpoint"})
    force_download: bool = field(default=False, metadata={"help": "Force the download of the model"})
    resume_download: bool = field(default=False, metadata={"help": "Resume an interrupted download"})
    proxies: Optional[str] = field(default=None, metadata={"help": "Proxy configuration for downloading the model"})
    output_loading_info: bool = field(default=False, metadata={"help": "Output additional loading information"})
    allow_remote_files: bool = field(default=False, metadata={"help": "Allow using remote files, default is local only"})
    revision: str = field(default="main", metadata={"help": "Specific model version to use"})
    trust_remote_code: bool = field(default=False, metadata={"help": "Enable trusting remote code when loading the model"})
    load_in_4bit: bool = field(default=False, metadata={"help": "Load model in 4-bit mode"})
    load_in_8bit: bool = field(default=False, metadata={"help": "Load model in 8-bit mode"})
    torch_dtype: Optional[str] = field(default=None, metadata={"help": "The torch dtype for the pre-trained model"})
    device_map: str = field(default="auto", metadata={"help": "The device map for the pre-trained model"})

    # Method to process additional arguments
    def process_additional_args(self, addt_args: List[str]):
        """
        Process additional cmd line args and update the model configuration accordingly.
        """
        addt_args_dict = {}
        i = 0
        while i < len(addt_args):
            key = addt_args[i].lstrip('-')  # Remove leading dashes
            if i + 1 < len(addt_args) and not addt_args[i + 1].startswith('--'):
                value = addt_args[i + 1]
                i += 2  # Move past the current key-value pair
            else:
                value = True  # Assign a True value for standalone flags
                i += 1  # Move to the next item

            addt_args_dict[key] = value

        # Update the ModelConfig instance with the additional args
        self.__dict__.update(addt_args_dict)

    def __post_init__(self): # validate parameters 
        """
        Post-initialization to validate some ModelConfig values
        """
        if self.torch_dtype and not hasattr(torch, self.torch_dtype):
            raise ValueError(f"Invalid torch dtype: {self.torch_dtype}")
        self.torch_dtype = getattr(torch, self.torch_dtype) if self.torch_dtype else None

        supported_pipelines = {"conversational", "text-generation"}
        if self.pipeline not in supported_pipelines:
            raise ValueError(f"Unsupported pipeline: {self.pipeline}")
        

parser = HfArgumentParser(ModelConfig)
args, additional_args = parser.parse_args_into_dataclasses(
    return_remaining_strings=True
)

args.process_additional_args(additional_args)

model_args = asdict(args)
model_args["local_files_only"] = not model_args.pop('allow_remote_files')
model_pipeline = model_args.pop('pipeline')
combination_type = model_args.pop('combination_type')

app = FastAPI()
tokenizer = AutoTokenizer.from_pretrained(**model_args)
base_model = AutoModelForCausalLM.from_pretrained(**model_args)

def list_files(directory):
    try:
        result = subprocess.run(['ls', directory], capture_output=True, text=True)
        if result.returncode == 0:
            return result.stdout.strip().split('\n')
        else:
            return [f"Command execution failed with return code: {result.returncode}"]
    except Exception as e:
        return [f"An error occurred: {str(e)}"]

output = list_files('/data')

adapters_list = [f"/data/{file}" for file in output]

if len(adapters_list) == 0:
    model = base_model
else: 
    model = PeftModel.from_pretrained(base_model, adapters_list[0], adapter_name="adapter-0")
    for i in range(1, len(adapters_list)):
        model.load_adapter(adapters_list[i], "adapter-"+str(i))
    adapters, weights= [], []
    for i in range(0, len(adapters_list)):
        adapters.append("adapter-"+str(i))
        adapter_num="ADAPTER_WEIGHT_"+str(i)
        weights.append(float(os.getenv(adapter_num)))

    model.add_weighted_adapter(
        adapters = adapters,
        weights = weights,
        adapter_name="combined_adapter",
        combination_type=combination_type,
    )
print("Model:",model)

pipeline_kwargs = {
    "trust_remote_code": args.trust_remote_code,
    "device_map": args.device_map,
}

if args.torch_dtype:
    pipeline_kwargs["torch_dtype"] = args.torch_dtype

pipeline = transformers.pipeline(
    model_pipeline,
    model=model,
    tokenizer=tokenizer,
    **pipeline_kwargs
)

try:
    # Attempt to load the generation configuration
    default_generate_config = GenerationConfig.from_pretrained(
        args.pretrained_model_name_or_path,
        local_files_only=args.local_files_only
    ).to_dict()
except Exception as e:
    default_generate_config = {}

class HomeResponse(BaseModel):
    message: str = Field(..., example="Server is running")
@app.get('/', response_model=HomeResponse, summary="Home Endpoint")
def home():
    """
    A simple endpoint that indicates the server is running.
    No parameters are required. Returns a message indicating the server status.
    """
    return {"message": "Server is running"}

class HealthStatus(BaseModel):
    status: str = Field(..., example="Healthy")
@app.get(
    "/healthz",
    response_model=HealthStatus,
    summary="Health Check Endpoint",
    responses={
        200: {
            "description": "Successful Response",
            "content": {
                "application/json": {
                    "example": {"status": "Healthy"}
                }
            }
        },
        500: {
            "description": "Error Response",
            "content": {
                "application/json": {
                    "examples": {
                        "model_uninitialized": {
                            "summary": "Model not initialized",
                            "value": {"detail": "Model not initialized"}
                        },
                        "pipeline_uninitialized": {
                            "summary": "Pipeline not initialized",
                            "value": {"detail": "Pipeline not initialized"}
                        }
                    }
                }
            }
        }
    }
)
def health_check():
    if not model:
        raise HTTPException(status_code=500, detail="Model not initialized")
    if not pipeline:
        raise HTTPException(status_code=500, detail="Pipeline not initialized")
    return {"status": "Healthy"}

class GenerateKwargs(BaseModel):
    max_length: int = 200 # Length of input prompt+max_new_tokens
    min_length: int = 0
    do_sample: bool = True
    early_stopping: bool = False
    num_beams: int = 1
    temperature: float = 1.0
    top_k: int = 10
    top_p: float = 1
    typical_p: float = 1
    repetition_penalty: float = 1
    pad_token_id: Optional[int] = tokenizer.pad_token_id
    eos_token_id: Optional[int] = tokenizer.eos_token_id
    class Config:
        extra = 'allow' # Allows for additional fields not explicitly defined
        json_schema_extra = {
            "example": {
                "max_length": 200,
                "temperature": 0.7,
                "top_p": 0.9,
                "additional_param": "Example value"
            }
        }

class Message(BaseModel):
    role: str
    content: str

class UnifiedRequestModel(BaseModel):
    # Fields for text generation
    prompt: Optional[str] = Field(None, description="Prompt for text generation. Required for text-generation pipeline. Do not use with 'messages'.")
    return_full_text: Optional[bool] = Field(True, description="Return full text if True, else only added text")
    clean_up_tokenization_spaces: Optional[bool] = Field(False, description="Clean up extra spaces in text output")
    prefix: Optional[str] = Field(None, description="Prefix added to prompt")
    handle_long_generation: Optional[str] = Field(None, description="Strategy to handle long generation")
    generate_kwargs: Optional[GenerateKwargs] = Field(default_factory=GenerateKwargs, description="Additional kwargs for generate method")

    # Field for conversational model
    messages: Optional[List[Message]] = Field(None, description="Messages for conversational model. Required for conversational pipeline. Do not use with 'prompt'.")
    def messages_to_dict_list(self):
        return [message.dict() for message in self.messages] if self.messages else []

class ErrorResponse(BaseModel):
    detail: str

@app.post(
    "/chat",
    summary="Chat Endpoint",
    responses={
        200: {
            "description": "Successful Response",
            "content": {
                "application/json": {
                    "examples": {
                        "text_generation": {
                            "summary": "Text Generation Response",
                            "value": {
                                "Result": "Generated text based on the prompt."
                            }
                        },
                        "conversation": {
                            "summary": "Conversation Response",
                            "value": {
                                "Result": "Response to the last message in the conversation."
                            }
                        }
                    }
                }
            }
        },
        400: {
            "model": ErrorResponse,
            "description": "Validation Error",
            "content": {
                "application/json": {
                    "examples": {
                        "missing_prompt": {
                            "summary": "Missing Prompt",
                            "value": {"detail": "Text generation parameter prompt required"}
                        },
                        "missing_messages": {
                            "summary": "Missing Messages",
                            "value": {"detail": "Conversational parameter messages required"}
                        }
                    }
                }
            }
        },
        500: {
            "model": ErrorResponse,
            "description": "Internal Server Error"
        }
    }
)
def generate_text(
        request_model: Annotated[
            UnifiedRequestModel,
            Body(
                openapi_examples={
                    "text_generation_example": {
                        "summary": "Text Generation Example",
                        "description": "An example of a text generation request.",
                        "value": {
                            "prompt": "Tell me a joke",
                            "return_full_text": True,
                            "clean_up_tokenization_spaces": False,
                            "prefix": None,
                            "handle_long_generation": None,
                            "generate_kwargs": GenerateKwargs().dict(),
                        },
                    },
                    "conversation_example": {
                        "summary": "Conversation Example",
                        "description": "An example of a conversational request.",
                        "value": {
                            "messages": [
                                {"role": "user", "content": "What is your favourite condiment?"},
                                {"role": "assistant", "content": "Well, im quite partial to a good squeeze of fresh lemon juice. It adds just the right amount of zesty flavour to whatever im cooking up in the kitchen!"},
                                {"role": "user", "content": "Do you have mayonnaise recipes?"}
                            ],
                            "return_full_text": True,
                            "clean_up_tokenization_spaces": False,
                            "prefix": None,
                            "handle_long_generation": None,
                            "generate_kwargs": GenerateKwargs().dict(),
                        },
                    },
                },
            ),
        ],
):
    """
    Processes chat requests, generating text based on the specified pipeline (text generation or conversational).
    Validates required parameters based on the pipeline and returns the generated text.
    """
    user_generate_kwargs = request_model.generate_kwargs.dict() if request_model.generate_kwargs else {}
    generate_kwargs = {**default_generate_config, **user_generate_kwargs}

    if args.pipeline == "text-generation":
        if not request_model.prompt:
            raise HTTPException(status_code=400, detail="Text generation parameter prompt required")
        sequences = pipeline(
            request_model.prompt,
            # return_tensors=request_model.return_tensors,
            # return_text=request_model.return_text,
            return_full_text=request_model.return_full_text,
            clean_up_tokenization_spaces=request_model.clean_up_tokenization_spaces,
            prefix=request_model.prefix,
            handle_long_generation=request_model.handle_long_generation,
            **generate_kwargs
        )

        result = ""
        for seq in sequences:
            print(f"Result: {seq['generated_text']}")
            result += seq['generated_text']

        return {"Result": result}

    elif args.pipeline == "conversational":
        if not request_model.messages:
            raise HTTPException(status_code=400, detail="Conversational parameter messages required")

        response = pipeline(
            request_model.messages_to_dict_list(),
            clean_up_tokenization_spaces=request_model.clean_up_tokenization_spaces,
            **generate_kwargs
        )
        return {"Result": str(response[-1])}

    else:
        raise HTTPException(status_code=400, detail="Invalid pipeline type")

class MemoryInfo(BaseModel):
    used: str
    total: str

class CPUInfo(BaseModel):
    load_percentage: float
    physical_cores: int
    total_cores: int
    memory: MemoryInfo

class GPUInfo(BaseModel):
    id: str
    name: str
    load: str
    temperature: str
    memory: MemoryInfo

class MetricsResponse(BaseModel):
    gpu_info: Optional[List[GPUInfo]] = None
    cpu_info: Optional[CPUInfo] = None

@app.get(
    "/metrics",
    response_model=MetricsResponse,
    summary="Metrics Endpoint",
    responses={
        200: {
            "description": "Successful Response",
            "content": {
                "application/json": {
                    "examples": {
                        "gpu_metrics": {
                            "summary": "Example when GPUs are available",
                            "value": {
                                "gpu_info": [{"id": "GPU-1234", "name": "GeForce GTX 950", "load": "25.00%", "temperature": "55 C", "memory": {"used": "1.00 GB", "total": "2.00 GB"}}],
                                "cpu_info": None  # Indicates CPUs info might not be present when GPUs are available
                            }
                        },
                        "cpu_metrics": {
                            "summary": "Example when only CPU is available",
                            "value": {
                                "gpu_info": None,  # Indicates GPU info might not be present when only CPU is available
                                "cpu_info": {"load_percentage": 20.0, "physical_cores": 4, "total_cores": 8, "memory": {"used": "4.00 GB", "total": "16.00 GB"}}
                            }
                        }
                    }
                }
            }
        },
        500: {
            "description": "Internal Server Error",
            "model": ErrorResponse,
        }
    }
)
def get_metrics():
    """
    Provides system metrics, including GPU details if available, or CPU and memory usage otherwise.
    Useful for monitoring the resource utilization of the server running the ML models.
    """
    try:
        if torch.cuda.is_available():
            gpus = GPUtil.getGPUs()
            gpu_info = [GPUInfo(
                id=gpu.id,
                name=gpu.name,
                load=f"{gpu.load * 100:.2f}%",
                temperature=f"{gpu.temperature} C",
                memory=MemoryInfo(
                    used=f"{gpu.memoryUsed / (1024 ** 3):.2f} GB",
                    total=f"{gpu.memoryTotal / (1024 ** 3):.2f} GB"
                )
            ) for gpu in gpus]
            return MetricsResponse(gpu_info=gpu_info)
        else:
            # Gather CPU metrics
            cpu_usage = psutil.cpu_percent(interval=1, percpu=False)
            physical_cores = psutil.cpu_count(logical=False)
            total_cores = psutil.cpu_count(logical=True)
            virtual_memory = psutil.virtual_memory()
            memory = MemoryInfo(
                used=f"{virtual_memory.used / (1024 ** 3):.2f} GB",
                total=f"{virtual_memory.total / (1024 ** 3):.2f} GB"
            )
            cpu_info = CPUInfo(
                load_percentage=cpu_usage,
                physical_cores=physical_cores,
                total_cores=total_cores,
                memory=memory
            )
            return MetricsResponse(cpu_info=cpu_info)
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

if __name__ == "__main__":
    local_rank = int(os.environ.get("LOCAL_RANK", 0)) # Default to 0 if not set
    port = 5000 + local_rank # Adjust port based on local rank
    uvicorn.run(app=app, host='0.0.0.0', port=port)