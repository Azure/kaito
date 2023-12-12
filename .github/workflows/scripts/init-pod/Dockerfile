FROM python:3.8-slim

WORKDIR /usr/src/app

# Install Git
RUN apt-get update && apt-get install -y git

# Copy the script
COPY main.py /usr/src/app

# Copy the build pod yaml
COPY build-pod.yaml /usr/src/app

# Set the entrypoint to the Python script
ENTRYPOINT ["python", "./main.py"]

