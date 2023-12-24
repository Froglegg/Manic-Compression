import logging
import azure.functions as func
import azure.durable_functions as df

async def main(msg: func.ServiceBusMessage, starter: str):
    client = df.DurableOrchestrationClient(starter)
    try:
        message_content = msg.get_body().decode("utf-8")
        instance_id = await client.start_new("ManicOrchestrator", None, message_content)
        logging.info(f"Started orchestration with ID = '{instance_id}', message body is `{message_content}`.")
    except Exception as e:
        logging.error(f"Error starting orchestration: {e}")