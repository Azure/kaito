import sys
import subprocess
from huggingface_hub import snapshot_download
from optimum.onnxruntime import AutoOptimizationConfig, ORTModelForCausalLM, ORTOptimizer

def download_and_convert_optimum_cli(repo_name): 
    # First try downloading from HF
    try:
        # Construct the optimum-cli command
        command = [
            "optimum-cli", "export", "onnx",
            "--model", repo_name, repo_name,
            "--task", "text-generation", "--framework", "pt",
            "--library", "transformers",
            "--cache_dir", repo_name
        ]

        subprocess.run(command, check=True)
        model = ORTModelForCausalLM.from_pretrained(repo_name, use_cache=False, use_io_binding=False)
    except Exception as e:
        try: 
            # Otherwise Manually download model
            download_path = snapshot_download(repo_id=repo_name)

            command = [
                "optimum-cli", "export", "onnx",
                "--model", download_path, repo_name,
                "--task", "text-generation", "--framework", "pt",
                "--library", "transformers",
                "--cache_dir", download_path
            ]

            subprocess.run(command, check=True)
            model = ORTModelForCausalLM.from_pretrained(repo_name, use_cache=False, use_io_binding=False, local_files_only=True)
        except Exception as e:
            print("Failed to convert model to ONNX using Optmum CLI", e)
            return None
    return model
    
def download_and_convert(repo_name):
    # Try converting to ONNX first with caching, if fails, retry with disabled caching mechanism
    # export=True flag specifies converting from pytorch to ONNX format
    try:
        model = ORTModelForCausalLM().from_pretrained(repo_name, export=True, provider="CUDAExecutionProvider")
    except Exception as e:
        try:
            model = ORTModelForCausalLM.from_pretrained(repo_name, use_cache=False, export=True, provider="CUDAExecutionProvider")
        # Lastly try using CLI
        except Exception as e: 
            print(f"Failed to load model with provider {provider}", e)
            model = download_and_convert_optimum_cli(repo_name)
    return model

def onnx_optimize_model(repo_name, model):
    try:
        optimizer = ORTOptimizer.from_pretrained(model)
        optimization_config = AutoOptimizationConfig.O2()
        optimizer.optimize(save_dir=repo_name, optimization_config=optimization_config)
    except NotImplementedError as e: 
        print("ONNX Runtime not supported for this model yet:", e)
    except Exception as e: 
        print("Optimizing model failed", e)

if __name__ == "__main__":
    repo_name = sys.argv[1]
    model = download_and_convert(repo_name)
    onnx_optimize_model(repo_name, model)

