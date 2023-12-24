import logging
import os
from azure.storage.blob import BlobClient
from os import path
from pedalboard import Pedalboard, Distortion
from pedalboard.io import AudioFile

def main(input) -> str:

    inputFile = input["inputFile"]
    sourceContainer = input["sourceContainer"]

    logging.info(f"Processing file {inputFile} in function {__name__}")

    # output_file (local only)
    outputFile = "effect2_" + inputFile

    # get properties
    storageConnectionString = os.environ["StorageConnectionString"]
    outputContainer = os.environ["OutputContainer"]

    # create client
    blob = BlobClient.from_connection_string(conn_str=storageConnectionString, container_name=sourceContainer, blob_name=inputFile)

    # download file
    with open('/tmp/' + inputFile, "wb") as my_blob:
        blob_data = blob.download_blob()
        blob_data.readinto(my_blob)

    # Make a Pedalboard object, containing multiple audio plugins:
    board = Pedalboard([Distortion(drive_db=40)])

    # Open an audio file for reading, just like a regular file:
    with AudioFile('/tmp/' + inputFile) as f:
  
        # Open an audio file to write to:
        with AudioFile('/tmp/' + outputFile, 'w', f.samplerate, f.num_channels) as o:
        
            # Read one second of audio at a time, until the file is empty:
            while f.tell() < f.frames:
                chunk = f.read(f.samplerate)
                
                # Run the audio through our pedalboard:
                effected = board(chunk, f.samplerate, reset=False)
                
                # Write the output to our output file:
                o.write(effected)

    # upload new file
    uploadBlob = BlobClient.from_connection_string(conn_str=storageConnectionString, container_name=outputContainer, blob_name=inputFile)
    with open('/tmp/' + outputFile, "rb") as data:
        uploadBlob.upload_blob(data=data, overwrite=True)

    # clean up
    os.remove('/tmp/' + inputFile)
    os.remove('/tmp/' + outputFile)

    return inputFile
