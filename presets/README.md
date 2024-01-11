# Kaito Preset Configurations
The current supported model families with preset configurations are listed below.

|Model Family |
|----|
|[falcon](./models/falcon)|
|[llama2](./models/llama2)|
|[llama2chat](./models/llama2chat)|


## Validation
Each preset model has its own hardware requirements in terms of GPU count and GPU memory defined in the respective `model.go` file. Kaito controller performs a validation check of whether the specified SKU and node count are sufficient to run the model or not. In case the provided SKU is not in the known list, the controller bypasses the validation check which means users need to ensure the model can run with the provided SKU. 

## Distributed inference

For models that support distributed inference, when the node count is larger than one, [torch distributed elastic](https://pytorch.org/docs/stable/distributed.elastic.html) is configured with master/worker pods running in multiple nodes and the service endpoint is the master pod.
