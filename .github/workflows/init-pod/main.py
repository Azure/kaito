import subprocess
import os
import shutil

KAITO_REPO_URL = "https://github.com/Azure/kaito.git"

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

    mod_files = check_modified_files(pr_branch)
    print("Modified files", mod_files)

    run_build_pods(mod_files, img_tag)

def run_build_pods(mod_files, img_tag):
    acr_name = os.environ.get("ACR_NAME", "aimodelsregistrytest")
    acr_username = os.environ.get("ACR_USERNAME")
    acr_password = os.environ.get("ACR_PASSWORD")

    for key, modified in mod_files.items(): 
        if modified:
            image_name = key.replace("_modified", "").replace("_", "-")

            job_yaml = populate_job_template(image_name, img_tag, acr_name, acr_username, acr_password)

            with open(f"{image_name}-job.yaml", "w") as file: 
                file.write(job_yaml)

            run_command(f"kubectl apply -f {image_name}-job.yaml")


def populate_job_template(image_name, img_tag, acr_name, acr_username, acr_password):
    with open("docker-job-template.yaml", "r") as file:
        job_template = file.read()

    # Replace placeholders with actual values
    job_template = job_template.replace("{{JOB_ID}}", image_name)
    job_template = job_template.replace("{{IMAGE_NAME}}", image_name)
    job_template = job_template.replace("{{IMAGE_TAG}}", img_tag)
    job_template = job_template.replace("{{ACR_NAME}}", acr_name)
    job_template = job_template.replace("{{ACR_USERNAME}}", acr_username)
    job_template = job_template.replace("{{ACR_PASSWORD}}", acr_password)

    return job_template

def check_modified_files(pr_branch):
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

    # Paths to check
    paths_to_check = {
        "falcon_7b_modified": "presets/falcon-7b/",
        "falcon_7b_chat_modified": "presets/falcon-7b-chat",
        "falcon_7b_chat_onnx_modified": "presets/falcon-7b-chat-onnx",
        "falcon_40b": "presets/falcon-40b",
        "falcon_40b_chat": "presets/falcon-40b-chat",
        "mistral_7b_v01": "presets/mistral-7b-v0.1",
        "mistral_7b_instruct_v01": "presets/mistral-7b-instruct-v0.1",
        "llama2_modified": "presets/llama-2/",
        "llama2_chat_modified": "presets/llama-2-chat/"
    }

    # Check modified files against paths
    modified_files = {key: path in files for key, path in paths_to_check.items()}

    # Print modified status
    for key, modified in modified_files.items():
        print(f"{key}: {modified}")

    return modified_files


if __name__ == "__main__":
    main()
