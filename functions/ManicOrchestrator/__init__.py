import logging
import json
import os
import azure.durable_functions as df
from azure.servicebus import ServiceBusClient, ServiceBusMessage

# get properties
serviceBusConnectionString = os.environ["AZURE_SERVICEBUS_CONNECTION_STRING"]

# constants
RESULTS_QUEUE_NAME = "audiotaskresults"

def send_result_to_queue(task):

    # create message
    msg = {
        "type": "processAudioResult",
        "content": json.dumps(task)
    }

    service_bus_client = ServiceBusClient.from_connection_string(serviceBusConnectionString)
    with service_bus_client:
        sender = service_bus_client.get_queue_sender(queue_name=RESULTS_QUEUE_NAME)
        with sender:
            results_message = ServiceBusMessage(json.dumps(msg))
            sender.send_messages(results_message)
            logging.info("Results sent to the results-queue.")


def orchestrator_function(context: df.DurableOrchestrationContext):
    message_content_raw = context.get_input()
    message = json.loads(message_content_raw)

    task = json.loads(message['content'])

    inputFile = task["inputFile"]
    audioFunctionPipeline = task['audioFunctionPipeline']

    current_input = inputFile

    inputContainer = os.environ["InputContainer"]
    outputContainer = os.environ["OutputContainer"]

    current_source_container = inputContainer

    # iterate over each function in the audioFunctionPipeline
    for function_name in audioFunctionPipeline:
        # call the activity function and pass the input file
        payload = {"inputFile": current_input, "sourceContainer": current_source_container}
        current_output = yield context.call_activity(function_name, payload)
        # set the current output as the input for the next activity function
        current_input = current_output
        current_source_container = outputContainer

    # set the task status to completed
    task["status"] = "Completed"
    task["outputFile"] = current_input

    # send the results to the results queue
    send_result_to_queue(task)

    return current_input

main = df.Orchestrator.create(orchestrator_function)
