import chainlit as cl
import requests
import os

URL = os.environ.get('WORKSPACE_SERVICE_URL')
@cl.step
def inference(prompt):
    # Endpoint URL
    data = {
        "prompt": prompt,
        "return_full_text": False,
        "clean_up_tokenization_spaces": True,
        "generate_kwargs": {
            "max_length": 1024
        }
    }

    response = requests.post(URL, json=data)

    if response.status_code == 200:
        response_data = response.json()
        return response_data.get("Result", "No result found")
    else:
        return f"Error: Received response code {response.status_code}"

@cl.on_message
async def main(message: cl.Message):
    """
    This function is called every time a user inputs a message in the UI.
    It sends back an intermediate response from inference, followed by the final answer.

    Args:
        message: The user's message.

    Returns:
        None.
    """

    # Call inference
    response = inference(message.content)

    # Send the final answer
    await cl.Message(content=response).send()