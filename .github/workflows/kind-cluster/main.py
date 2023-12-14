import subprocess
import os
import shutil
from pathlib import Path
import time
import random
import string

KAITO_REPO_URL = "https://github.com/Azure/kaito.git"

HOST_WEIGHTS_PATHS = {
    # TFS Weights
    "falcon-7b": "/home/tfs/tiiuae/falcon-7b/weights",
    "falcon-7b-instruct": "/home/tfs/tiiuae/falcon-7b-instruct/weights",
    "falcon-40b": "/home/tfs/tiiuae/falcon-40b/weights",
    "falcon-40b-instruct": "/home/tfs/tiiuae/falcon-40b-instruct/weights",
    "mistral-7b-v01": "/home/tfs/mistralai/mistral-7b-v0.1/weights",
    "mistral-7b-instruct-v0.1": "/home/tfs/mistralai/mistral-7b-instruct-v0.1/weights",

    # TFS Onnx Weights
    "falcon-7b-instruct-onnx": "/home/tfs/tiiuae/falcon-7b-instruct-onnx/weights",

    # Llama Weights (Mounted on /datadrive drive)
    "llama-2-7b": "/datadrive/llama/llama-2-7b",
    "llama-2-7b-chat": "/datadrive/llama/llama-2-7b-chat",
    "llama-2-13b": "/datadrive/llama/llama-2-13b",
    "llama-2-13b-chat": "/datadrive/llama/llama-2-13b-chat",
    "llama-2-70b": "/datadrive/llama/llama-2-70b",
    "llama-2-70b-chat": "/datadrive/llama/llama-2-70b-chat"
}

REPO_PRESET_PATHS = {
    # Falcon Presets
    "falcon-7b": "/kaito/presets/models/falcon",
    "falcon-7b-instruct": "/kaito/presets/models/falcon",
    "falcon-40b": "/kaito/presets/models/falcon",
    "falcon-40b-instruct": "/kaito/presets/models/falcon",

    # Mistral Presets
    "mistral-7b-v01": "/kaito/presets/models/mistral",
    "mistral-7b-instruct-v0.1": "/kaito/presets/models/mistral",

    # TFS Onnx Presets
    "falcon-7b-instruct-onnx": "/kaito/presets/models/falcon",

    # Llama Presets
    "llama-2-7b": "/kaito/presets/models/llama2",
    "llama-2-7b-chat": "/kaito/presets/models/llama2chat",
    "llama-2-13b": "/kaito/presets/models/llama2",
    "llama-2-13b-chat": "/kaito/presets/models/llama2chat",
    "llama-2-70b": "/kaito/presets/models/llama2",
    "llama-2-70b-chat": "/kaito/presets/models/llama2chat"
}


REPO_DOCKERFILE_PATHS = {
    # TFS Presets
    "falcon-7b": "/kaito/docker/presets/tfs/Dockerfile",
    "falcon-7b-instruct": "/kaito/docker/presets/tfs/Dockerfile",
    "falcon-40b": "/kaito/docker/presets/tfs/Dockerfile",
    "falcon-40b-instruct": "/kaito/docker/presets/tfs/Dockerfile",
    "mistral-7b-v01": "/kaito/docker/presets/tfs/Dockerfile",
    "mistral-7b-instruct-v0.1": "/kaito/docker/presets/tfs/Dockerfile",

    # TFS Onnx Presets
    "falcon-7b-instruct-onnx": "/kaito/docker/presets/tfs-onnx/Dockerfile",

    # Llama Presets
    "llama-2-7b": "/kaito/docker/presets/llama-2/Dockerfile",
    "llama-2-7b-chat": "/kaito/docker/presets/llama-2/Dockerfile",
    "llama-2-13b": "/kaito/docker/presets/llama-2/Dockerfile",
    "llama-2-13b-chat": "/kaito/docker/presets/llama-2/Dockerfile",
    "llama-2-70b": "/kaito/docker/presets/llama-2/Dockerfile",
    "llama-2-70b-chat": "/kaito/docker/presets/llama-2/Dockerfile"
}

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
        with open("docker-job-template.yaml", "r") as file:
            job_template = file.read()

        replacements = {
            "{{JOB_ID}}": f"{job_name}",
            "{{IMAGE_NAME}}": model,
            "{{IMAGE_TAG}}": img_tag,
            "{{ACR_NAME}}": env_vars["ACR_NAME"],
            "{{ACR_USERNAME}}": env_vars["ACR_USERNAME"],
            "{{ACR_PASSWORD}}": env_vars["ACR_PASSWORD"],
            "{{PR_BRANCH}}": env_vars["PR_BRANCH"],
            "{{HOST_WEIGHTS_PATH}}": HOST_WEIGHTS_PATHS[model],
            "{{MODEL_PRESET_PATH}}": REPO_PRESET_PATHS[model],
            "{{DOCKERFILE_PATH}}": REPO_DOCKERFILE_PATHS[model]
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

    modified_models = {model: preset_path in files for model, preset_path in REPO_PRESET_PATHS.items()}

    return modified_models

def check_job_status(job_name):
    """Check the status of a Kubernetes job."""
    # Query for the specific fields 'succeeded' and 'failed' in the job's status
    command_succeeded = f"kubectl get job docker-build-job-{job_name} -o jsonpath='{{.status.succeeded}}'"
    command_failed = f"kubectl get job docker-build-job-{job_name} -o jsonpath='{{.status.failed}}'"

    succeeded = run_command(command_succeeded)
    failed = run_command(command_failed)

    if succeeded and int(succeeded) > 0:
        return "succeeded"
    elif failed and int(failed) > 0:
        return "failed"
    else:
        return "running"

def wait_for_jobs_to_complete(job_names, timeout=3600):
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
        if all_completed:
            print("All jobs completed successfully.")
            return True
        time.sleep(30)  # Wait for 30 seconds before checking again
    print("Timeout waiting for jobs to complete.")
    return False

if __name__ == "__main__":
    main()
