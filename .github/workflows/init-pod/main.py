import subprocess
import os
import shutil

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
    "falcon-7b": "kaito/presets/models/falcon",
    "falcon-7b-instruct": "kaito/presets/models/falcon",
    "falcon-40b": "kaito/presets/models/falcon",
    "falcon-40b-instruct": "kaito/presets/models/falcon",

    # Mistral Presets
    "mistral-7b-v01": "kaito/presets/models/mistral",
    "mistral-7b-instruct-v0.1": "kaito/presets/models/mistral",

    # TFS Onnx Presets
    "falcon-7b-instruct-onnx": "kaito/presets/models/falcon",

    # Llama Presets
    "llama-2-7b": "kaito/presets/models/llama2",
    "llama-2-7b-chat": "kaito/presets/models/llama2chat",
    "llama-2-13b": "kaito/presets/models/llama2",
    "llama-2-13b-chat": "kaito/presets/models/llama2chat",
    "llama-2-70b": "kaito/presets/models/llama2",
    "llama-2-70b-chat": "kaito/presets/models/llama2chat"
}


REPO_DOCKERFILE_PATHS = {
    # TFS Presets
    "falcon-7b": "kaito/docker/presets/tfs/Dockerfile",
    "falcon-7b-instruct": "kaito/docker/presets/tfs/Dockerfile",
    "falcon-40b": "kaito/docker/presets/tfs/Dockerfile",
    "falcon-40b-instruct": "kaito/docker/presets/tfs/Dockerfile",
    "mistral-7b-v01": "kaito/docker/presets/tfs/Dockerfile",
    "mistral-7b-instruct-v0.1": "kaito/docker/presets/tfs/Dockerfile",

    # TFS Onnx Presets
    "falcon-7b-instruct-onnx": "kaito/docker/presets/tfs-onnx/Dockerfile",

    # Llama Presets
    "llama-2-7b": "kaito/docker/presets/llama-2/Dockerfile",
    "llama-2-7b-chat": "kaito/docker/presets/llama-2/Dockerfile",
    "llama-2-13b": "kaito/docker/presets/llama-2/Dockerfile",
    "llama-2-13b-chat": "kaito/docker/presets/llama-2/Dockerfile",
    "llama-2-70b": "kaito/docker/presets/llama-2/Dockerfile",
    "llama-2-70b-chat": "kaito/docker/presets/llama-2/Dockerfile"
}

def run_command(command):
    process = subprocess.Popen(command, stdout=subprocess.PIPE, shell=True)
    output, error = process.communicate()
    if error:
        print(f"Error: {error}")
    return output.decode('utf-8').strip()

def main(): 
    pr_branch = os.environ.get("PR_BRANCH", "main")
    print(f"pr_branch: {pr_branch}")

    img_tag = os.environ.get("IMAGE_TAG", "0.0.1")
    print(f"image_tag: {img_tag}")

    mod_models = check_modified_models(pr_branch)
    print("Modified files", mod_models)

    run_build_pods(pr_branch, img_tag, mod_models)

def run_build_pods(pr_branch, img_tag, mod_models):
    acr_name = os.environ.get("ACR_NAME", "aimodelsregistrytest")
    acr_username = os.environ.get("ACR_USERNAME")
    acr_password = os.environ.get("ACR_PASSWORD")

    for model, modified in mod_models.items(): 
        if modified:
            image_name = model
            job_yaml = populate_job_template(image_name, img_tag, acr_name, acr_username, acr_password, pr_branch)

            with open(f"{image_name}-job.yaml", "w") as file: 
                file.write(job_yaml)

            run_command(f"kubectl apply -f {image_name}-job.yaml")


def populate_job_template(image_name, img_tag, acr_name, acr_username, acr_password, pr_branch):
    with open("docker-job-template.yaml", "r") as file:
        job_template = file.read()

    # Replace placeholders with actual values
    job_template = job_template.replace("{{JOB_ID}}", image_name)
    job_template = job_template.replace("{{IMAGE_NAME}}", image_name)
    job_template = job_template.replace("{{IMAGE_TAG}}", img_tag)
    job_template = job_template.replace("{{ACR_NAME}}", acr_name)
    job_template = job_template.replace("{{ACR_USERNAME}}", acr_username)
    job_template = job_template.replace("{{ACR_PASSWORD}}", acr_password)
    job_template = job_template.replace("{{PR_BRANCH}}", pr_branch)
    job_template = job_template.replace("{{HOST_WEIGHTS_PATH}}", HOST_WEIGHTS_PATHS[image_name])
    job_template = job_template.replace("{{MODEL_PRESET_PATH}}", REPO_PRESET_PATHS[image_name])
    job_template = job_template.replace("{{DOCKERFILE_PATH}}", REPO_DOCKERFILE_PATHS[image_name])

    return job_template

def check_modified_models(pr_branch):
    repo_dir = "repo"
    repo_path = os.path.join(os.getcwd(), repo_dir)

    # Ensure the repo directory is clean before starting
    if os.path.exists(repo_path):
        shutil.rmtree(repo_path)

    # Clone the repo
    run_command(f"git clone {KAITO_REPO_URL} {repo_dir}")
    os.chdir(repo_dir)

    # Setup for fetching
    run_command("git checkout --detach")
    run_command("git fetch origin main:main")
    run_command(f"git fetch origin {pr_branch}:{pr_branch}")

    # Checkout the PR branch
    run_command(f"git checkout {pr_branch}")

    # Get modified files
    files = run_command("git diff --name-only origin/main")

    # Reset back to home directory
    os.chdir("..")

    print("Modified files:", files)

    # Check modified models against paths
    modified_models = {model: preset_path in files for model, preset_path in REPO_PRESET_PATHS.items()}

    # Print modified status
    for key, modified in modified_models.items():
        print(f"{key}: {modified}")

    modified_models["falcon-7b"] = True # FOR TESTING

    return modified_models


if __name__ == "__main__":
    main()
