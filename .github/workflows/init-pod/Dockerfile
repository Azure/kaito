FROM python:3.8-slim

WORKDIR /usr/src/app

# Install Git
RUN apt-get update && apt-get install -y git

# Copy the script
COPY main.py /usr/src/app

# Copy the docker job yaml
COPY docker-job-template.yaml /usr/src/app

# Set the entrypoint to the Python script
ENTRYPOINT ["python3", "main.py"]

