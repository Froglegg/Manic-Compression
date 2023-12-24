import logging
import os
from azure.storage.blob import BlobClient
from os import path
from pydub import AudioSegment

def main(input) -> str:

    inputFile = input["inputFile"]
    sourceContainer = input["sourceContainer"]

    logging.info(f"Processing file {inputFile} in function {__name__}")
    # this task creates a new file
    outputFile = inputFile.replace("wav", "mp3")

    # get properties
    storageConnectionString = os.environ["StorageConnectionString"]
    outputContainer = os.environ["OutputContainer"]

    # create client
    blob = BlobClient.from_connection_string(conn_str=storageConnectionString, container_name=sourceContainer, blob_name=inputFile)

    # download file
    with open('/tmp/' + inputFile, "wb") as my_blob:
        blob_data = blob.download_blob()
        blob_data.readinto(my_blob)

    # convert
    sound = AudioSegment.from_mp3('/tmp/' + inputFile)
    sound.export('/tmp/' + outputFile, format="wav")

    # cleanup source if we are on the out container
    if sourceContainer == outputContainer:
        # delete to prevent multiple output files
        blob.delete_blob()

    # upload new file
    uploadBlob = BlobClient.from_connection_string(conn_str=storageConnectionString, container_name=outputContainer, blob_name=outputFile)
    with open('/tmp/' + outputFile, "rb") as data:
        uploadBlob.upload_blob(data=data, overwrite=True)

    # clean up
    os.remove('/tmp/' + inputFile)
    os.remove('/tmp/' + outputFile)

    return outputFile
