import sys
import subprocess
from optimum.onnxruntime import AutoOptimizationConfig, ORTModelForCausalLM, ORTOptimizer
import glob

def get_model_name(repo_name): 
    last_slash_idx= repo_name.rfind('/')
    model_name = repo_name[last_slash_index + 1:] if last_slash_index != -1 else repo_name
    return model_name

def download_and_convert(repo_name, model_name):
    """
    Download and convert a model to ONNX format.
    
    Parameters:
    repo_name (str): The repository name to download the model from.
    model_name (str): The name to save the ONNX model as.

    Returns:
    ORTModelForCausalLM: The loaded or converted ONNX model, or None if failed.
    """
    if not repo_name or not model_name:
        print("Repository name and model name must be provided")
        return None

    # Search for .onnx files in the current and subdirectories
    onnx_files = glob.glob('**/*.onnx', recursive=True)

    # If an ONNX file is found, use it
    if onnx_files:
        file_path = onnx_files[0]
        try:
            model = ORTModelForCausalLM.from_pretrained(file_path)
            print(f"Loaded local ONNX model from {file_path}")
            return model
        except Exception as e:
            print(f"Loading local ONNX model from {file_path} failed: {e}")

    # Try converting to ONNX with caching, then without if fails
    try:
        model = ORTModelForCausalLM.from_pretrained(repo_name, export=True, provider="CUDAExecutionProvider")
        model.save_pretrained(f"{model_name}.onnx")
        print(f"Model converted and saved as {model_name}.onnx")
        return model
    except Exception as e:
        print(f"Failed to convert model to ONNX with caching: {e}")
    
    try:
        model = ORTModelForCausalLM.from_pretrained(repo_name, use_cache=False, export=True, provider="CUDAExecutionProvider")
        model.save_pretrained(f"{model_name}.onnx")
        print(f"Model converted without cache and saved as {model_name}.onnx")
        return model
    except Exception as e:
        print(f"Failed to convert model from {repo_name} without caching: {e}")
    return None

def onnx_optimize_model(model, model_name):
    try:
        optimizer = ORTOptimizer.from_pretrained(model)
        optimization_config = AutoOptimizationConfig.O2()
        optimizer.optimize(save_dir=model_name, optimization_config=optimization_config)
    except NotImplementedError as e: 
        print("ONNX Optimization not supported for this model yet:", e)
    except Exception as e: 
        print("Optimizing model failed", e)

if __name__ == "__main__":
    repo_name = sys.argv[1]
    model_name = get_model_name(repo_name)
    model = download_and_convert(repo_name, model_name)
    if model: 
        onnx_optimize_model(model, model_name)

