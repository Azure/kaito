FROM python:3.12-slim

WORKDIR /app

# Copy all files from ragengine/services into the app/services folder
COPY presets/ragengine/ ragengine/

# Set the PYTHONPATH environment variable
ENV PYTHONPATH=/app

# Install dependencies from requirements.txt
RUN pip install --no-cache-dir -r ragengine/requirements.txt

# Set the final working directory
WORKDIR /app/ragengine

# Run Using Following
# docker build -f docker/ragengine/service/Dockerfile -t kaito-rag-service:0.0.1 .
# docker tag kaito-rag-service:0.0.1 <ACR_PATH>/kaito-rag-service:0.0.1
# docker push <ACR_PATH>/kaito-rag-service:0.0.1

