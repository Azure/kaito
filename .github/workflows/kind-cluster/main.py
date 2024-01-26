import os
import random
import shutil
import string
import subprocess
import time
from pathlib import Path

import yaml

KAITO_REPO_URL = "https://github.com/Azure/kaito.git"

def get_weights_path(model_name): 
    return f"/datadrive/{model_name}/weights"

def get_dockerfile_path(model_runtime): 
    return f"/kaito/docker/presets/{model_runtime}/Dockerfile"

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

def update_model(model_name, model_version): 
    """Using Git Update Model"""
    if not model_version: 
        return
    weights_path = get_weights_path(model_name)
    start_dir = os.getcwd()
    try:
        # Change to weights directory 
        os.chdir(weights_path)
        run_command("git checkout main")
        run_command("git pull origin main")
        run_command(f"git checkout {model_version}")
    except Exception as e:
        print(f"An error occurred: {e}")
    finally: 
        # Change back to the original directory
        os.chdir(start_dir)

def download_new_model(model_name, model_url):
    """Given URL download new model""" 
    weights_path = get_weights_path(model_name)
    start_dir = os.getcwd()
    # If a new model then download it
    if not os.path.exists(weights_path) and model_url: 
        try:
            os.makedirs(weights_path, exist_ok=True)
            # Change to weights directory 
            os.chdir(weights_path)
            # Clone the repo
            run_command(f"git clone {model_url}")
        except Exception as e:
            print(f"An error occurred: {e}")
        finally: 
            # Change back to the original directory
            os.chdir(start_dir)

def main():
    pr_branch = os.environ.get("PR_BRANCH", "main")
    model_name = os.environ.get("MODEL_NAME", None)
    model_version = os.environ.get("MODEL_VERSION", None)
    model_runtime = os.environ.get("MODEL_RUNTIME", None)
    model_type = os.environ.get("MODEL_TYPE", None)
    model_tag = os.environ.get("MODEL_TAG", None)
    model_url = os.environ.get("MODEL_URL", None)

    download_new_model(model_name, model_url)
    update_model(model_name, model_version)
    clone_and_checkout_pr_branch(pr_branch)

    job_names = []

    unique_id = generate_unique_id()
    job_name = f"{model_name}-{unique_id}"
    job_yaml = populate_job_template(model_name, model_type, model_runtime, model_tag, job_name, os.environ)
    write_job_file(job_yaml, job_name)

    output = run_command(f"ls {get_weights_path(model_name)}")
    print("Model Weights:", output)

    run_command(f"kubectl apply -f {job_name}-job.yaml")
    job_names.append(job_name)
    
    if not wait_for_jobs_to_complete(job_names):
        exit(1)  # Exit with an error code if any job failed

def write_job_file(job_yaml, job_name):
    """Write the job yaml to a file."""
    if job_yaml:
        with open(f"{job_name}-job.yaml", "w") as file:
            file.write(job_yaml)

def clone_and_checkout_pr_branch(pr_branch):
    """Clone and checkout PR Branch."""
    repo_dir = Path.cwd() / "repo"

    if repo_dir.exists():
        shutil.rmtree(repo_dir)

    run_command(f"git clone {KAITO_REPO_URL} {repo_dir}")
    os.chdir(repo_dir)

    run_command("git checkout --detach")
    run_command("git fetch origin main:main")
    run_command(f"git fetch origin {pr_branch}:{pr_branch}")
    run_command(f"git checkout {pr_branch}")

    os.chdir(Path.cwd().parent)

def populate_job_template(model_name, model_type, model_runtime, model_tag, job_name, env_vars):
    """Populate the job template with provided values."""
    try:
        docker_job_template = Path.cwd() / "repo/.github/workflows/kind-cluster/docker-job-template.yaml"
        with open(docker_job_template, "r") as file:
            job_template = file.read()

        replacements = {
            "{{JOB_ID}}": f"{job_name}",
            "{{IMAGE_NAME}}": model_name,
            "{{ACR_NAME}}": env_vars["ACR_NAME"],
            "{{ACR_USERNAME}}": env_vars["ACR_USERNAME"],
            "{{ACR_PASSWORD}}": env_vars["ACR_PASSWORD"],
            "{{PR_BRANCH}}": env_vars["PR_BRANCH"],
            "{{HOST_WEIGHTS_PATH}}": get_weights_path(model_name),
            "{{MODEL_TYPE}}": model_type,
            "{{DOCKERFILE_PATH}}": get_dockerfile_path(model_runtime),
            "{{VERSION}}": model_tag,
        }

        for key, value in replacements.items():
            job_template = job_template.replace(key, value)

        return job_template
    except Exception as e:
        print(f"An error occurred while populating job template: {e}")
        return None


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

def wait_for_jobs_to_complete(job_names, timeout=21600):
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
