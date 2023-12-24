import React, { useEffect, useState } from "react";
import {
  ChakraProvider,
  Button,
  Flex,
  VStack,
  Grid,
  GridItem,
  Container,
  Divider,
} from "@chakra-ui/react";

import { AudioFunctionSelector, FileTable, TaskStatus } from "./views";
import { Banner, LoadingSpinner, ModalButton } from "./components";
import { getEndpoint, handleStart } from "./api";

const App = () => {
  const [isLoading, setIsLoading] = useState(false);
  const [inputFiles, setInputFiles] = useState([]);
  const [outputFiles, setOutputFiles] = useState([]);

  const [functions, setFunctions] = useState([]);
  const [selectedAudioFunctions, setSelectedAudioFunctions] = useState([]);
  const [audioFunctionPipeline, setAudioFunctionPipeline] = useState([]);

  useEffect(() => {
    setIsLoading(true);
    const initialize = async () => {
      try {
        const [inputFiles, outputFiles, functions] = await Promise.all([
          getEndpoint("/input/"),
          getEndpoint("/output/"),
          getEndpoint("/functions/"),
        ]);
        setInputFiles(inputFiles);
        setOutputFiles(outputFiles);
        setFunctions(functions);
      } catch (error) {
        console.error("Failed to initialize:", error);
      } finally {
        setIsLoading(false);
      }
    };

    initialize();
  }, []);

  const handleStartManicCompression = async () => {
    setIsLoading(true);
    setOutputFiles([]);
    try {
      const job = {
        inputFiles: inputFiles.map((file) => file.Name),
        clientID: "webclient:3000",
        audioFunctionPipeline: audioFunctionPipeline,
      };
      await handleStart(job);
    } catch (error) {
      console.error("Failed to start manic compression:", error);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <ChakraProvider>
      <Banner heading={"Manic Compression"} />
      <Container maxW="container.xl" p={4}>
        {isLoading && <LoadingSpinner />}
        <VStack spacing={6}>
          <Grid
            templateColumns={{ sm: "1fr 1fr 1fr", md: "1fr 1fr 1fr" }}
            gap={6}
            width="full"
          >
            <AudioFunctionSelector
              functions={functions}
              selectedAudioFunctions={selectedAudioFunctions}
              setSelectedAudioFunctions={setSelectedAudioFunctions}
              setAudioFunctionPipeline={setAudioFunctionPipeline} // new prop
            />
          </Grid>
          <Flex justifyContent="flex-end" width="full">
            <ModalButton
              modalButtonText="Poll Active Tasks"
              modalHeaderText="Active Tasks"
              size="full"
              mx={4}
            >
              <TaskStatus queue="active" />
            </ModalButton>
            <ModalButton
              modalButtonText="Poll Completed Tasks"
              modalHeaderText="Completed Tasks"
              size="full"
              mx={4}
            >
              <TaskStatus queue="completed" />
            </ModalButton>
            <Button
              colorScheme="teal"
              onClick={() => handleStartManicCompression()}
              isDisabled={
                selectedAudioFunctions.length === 0 ||
                isLoading ||
                inputFiles.length === 0 ||
                audioFunctionPipeline.length === 0
              }
              mx={4}
            >
              Process Audio Input
            </Button>
          </Flex>
          <Divider />
          <Grid
            templateColumns={{ sm: "1fr", md: "1fr 1fr" }}
            gap={6}
            width="full"
          >
            <GridItem>
              <FileTable
                files={inputFiles}
                setIsLoading={setIsLoading}
                setFiles={setInputFiles}
                containerPath={"/input/"}
                header={"Input Files"}
              />
            </GridItem>
            <GridItem>
              <FileTable
                files={outputFiles}
                setIsLoading={setIsLoading}
                setFiles={setOutputFiles}
                containerPath={"/output/"}
                header={"Output Files"}
                showUploadFileButton={false}
              />
            </GridItem>
          </Grid>
        </VStack>
      </Container>
    </ChakraProvider>
  );
};

export default App;
