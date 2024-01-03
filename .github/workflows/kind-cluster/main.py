import subprocess
import os
import shutil
from pathlib import Path
import time
import random
import string
import yaml

KAITO_REPO_URL = "https://github.com/Azure/kaito.git"

def read_models_from_yaml(file_path):
    with open(file_path, 'r') as file:
        data = yaml.safe_load(file)
        return set(data['models'])

yaml_file_path = 'presets/models/supported_models.yaml'
MODELS = read_models_from_yaml(yaml_file_path)

def get_model_type(model_name): 
    model_type = "tfs"
    if "llama" in model_name: 
        model_type = "llama-2"
    elif "onnx" in model_name: 
        model_type = "tfs-onnx"
    return model_type

def get_weights_path(model_name): 
    return f"/datadrive/{model_name}/weights"

def get_preset_path(model_name): 
    preset_name = model_name.split("-")[0]
    if preset_name == "llama":
        preset_name += "2"
        if model_name.endswith("chat"):
            preset_name += "chat"
    return f"/kaito/presets/models/{preset_name}"

def get_dockerfile_path(model_name): 
    model_type = get_model_type(model_name)
    return f"/kaito/docker/presets/{model_type}/Dockerfile"

def generate_unique_id():
    """Generate a unique identifier for a job."""
    timestamp = int(time.time())
    random_str = ''.join(random.choices(string.ascii_lowercase + string.digits, k=6))
    return f"{timestamp}-{random_str}"

def run_command(command):
    """Execute a shell command and return the output."""
    try:
        process = subprocess.Popen(command, stdout=subprocess.PIPE, shell=True)
        output, error = process.communicate()
        if error:
            print(f"Error: {error}")
        return output.decode('utf-8').strip()
    except Exception as e:
        print(f"An error occurred: {e}")
        return None

def main(): 
    pr_branch = os.environ.get("PR_BRANCH", "main")
    img_tag = os.environ.get("IMAGE_TAG", "0.0.1")
    mod_models = check_modified_models(pr_branch)

    job_names = []
    for model, modified in mod_models.items(): 
        if modified:
            unique_id = generate_unique_id()
            job_name = f"{model}-{unique_id}"
            job_yaml = populate_job_template(model, img_tag, job_name, os.environ)
            write_job_file(job_yaml, job_name)
            run_command(f"kubectl apply -f {job_name}-job.yaml")
            job_names.append(job_name)
    
    if not wait_for_jobs_to_complete(job_names):
        exit(1)  # Exit with an error code if any job failed

def write_job_file(job_yaml, job_name):
    """Write the job yaml to a file."""
    if job_yaml:
        with open(f"{job_name}-job.yaml", "w") as file:
            file.write(job_yaml)

def populate_job_template(model, img_tag, job_name, env_vars):
    """Populate the job template with provided values."""
    try:
        with open("/home/azureuser/docker-job-template.yaml", "r") as file:
            job_template = file.read()

        replacements = {
            "{{JOB_ID}}": f"{job_name}",
            "{{IMAGE_NAME}}": model,
            "{{IMAGE_TAG}}": img_tag,
            "{{ACR_NAME}}": env_vars["ACR_NAME"],
            "{{ACR_USERNAME}}": env_vars["ACR_USERNAME"],
            "{{ACR_PASSWORD}}": env_vars["ACR_PASSWORD"],
            "{{PR_BRANCH}}": env_vars["PR_BRANCH"],
            "{{HOST_WEIGHTS_PATH}}": get_weights_path(model),
            "{{MODEL_PRESET_PATH}}": get_preset_path(model),
            "{{DOCKERFILE_PATH}}": get_dockerfile_path(model)
        }

        for key, value in replacements.items():
            job_template = job_template.replace(key, value)

        return job_template
    except Exception as e:
        print(f"An error occurred while populating job template: {e}")
        return None

def check_modified_models(pr_branch):
    """Check for modified models in the repository."""
    repo_dir = Path.cwd() / "repo"

    if repo_dir.exists():
        shutil.rmtree(repo_dir)

    run_command(f"git clone {KAITO_REPO_URL} {repo_dir}")
    os.chdir(repo_dir)

    run_command("git checkout --detach")
    run_command("git fetch origin main:main")
    run_command(f"git fetch origin {pr_branch}:{pr_branch}")
    run_command(f"git checkout {pr_branch}")

    files = run_command("git diff --name-only origin/main")
    os.chdir(Path.cwd().parent)

    modified_models = {model: model.split("-")[0] in files for model in MODELS}
    print("Modified Models (Images to build): ", modified_models)

    return modified_models

def check_job_status(job_name):
    """Check the status of a Kubernetes job."""
    # Query for both 'succeeded' and 'failed' fields in the job's status
    command = f"kubectl get job docker-build-job-{job_name} -o jsonpath='{{.status.succeeded}} {{.status.failed}}'"

    status = run_command(command).split()
    
    # Check if status list has two elements (succeeded and failed)
    if len(status) == 2:
        succeeded, failed = status
        if succeeded and int(succeeded) > 0:
            return "succeeded"
        elif failed and int(failed) > 0:
            return "failed"
        else:
            return "running"
    else:
        return "unknown"  # This case handles situations where the job status might not be available


def wait_for_jobs_to_complete(job_names, timeout=10800):
    """Wait for all jobs to complete with a timeout."""
    start_time = time.time()
    while time.time() - start_time < timeout:
        all_completed = True
        for job_name in job_names:
            print("Check Job Status: ", job_name)
            status = check_job_status(job_name)
            if status != "succeeded":
                all_completed = False
                if status == "failed":
                    print(f"Job {job_name} failed.")
                    return False
            time.sleep(5) # Wait for 5 sec between requests - prevents connection errors
        if all_completed:
            print("All jobs completed successfully.")
            return True
        time.sleep(30)  # Wait for 30 seconds before checking again
    print("Timeout waiting for jobs to complete.")
    return False

if __name__ == "__main__":
    main()
