FROM python:3.12-slim

WORKDIR /app

# Copy all files from ragengine/services into the app/services folder
COPY services/ services/

# Install dependencies from requirements.txt
RUN pip install --no-cache-dir -r services/requirements.txt

# Set the final working directory
WORKDIR /app/services

# Run Using Following
# docker build -f kaito-repo/pkg/ragengine/docker/Dockerfile.service -t kaito-rag-service:0.0.1 kaito-repo/pkg/ragengine
# docker tag kaito-rag-service:0.0.1 <ACR_PATH>/kaito-rag-service:0.0.1
# docker push <ACR_PATH>/kaito-rag-service:0.0.1

