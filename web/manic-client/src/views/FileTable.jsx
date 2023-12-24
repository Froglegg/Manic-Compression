import React, { useState, useEffect } from "react";
import { Box, List, Heading, Input, Button, Stack } from "@chakra-ui/react";
import {
  PlusSquareIcon,
  DeleteIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  RepeatIcon,
} from "@chakra-ui/icons";
import { FileListItem } from "../components/index";

import {
  handleDelete,
  handleFileDownload,
  handleFileUpload,
  getEndpoint,
  handleContainerClear,
} from "../api";

const FileTable = ({
  containerPath,
  header,
  files,
  setFiles,
  setIsLoading,
  showUploadFileButton = true,
}) => {
  // pagination
  const [currentPage, setCurrentPage] = useState(1);
  const itemsPerPage = 15;

  useEffect(() => {
    setCurrentPage(1);
  }, [files]);

  const paginate = (pageNumber) => setCurrentPage(pageNumber);

  const totalPages = Math.ceil(files.length / itemsPerPage);
  const indexOfLastItem = currentPage * itemsPerPage;
  const indexOfFirstItem = indexOfLastItem - itemsPerPage;
  const currentFiles = files.slice(indexOfFirstItem, indexOfLastItem);

  const onDownload = async (fileName) => {
    setIsLoading(true);
    try {
      handleFileDownload(containerPath, fileName);
    } catch (error) {
      console.error("Failed to download file:", error);
    } finally {
      setIsLoading(false);
    }
  };

  const onRefresh = async () => {
    setIsLoading(true);
    try {
      const files = await getEndpoint(containerPath);
      setFiles(files);
    } catch (error) {
      console.error("Failed to refresh files:", error);
    } finally {
      setIsLoading(false);
    }
  };

  const onDelete = async (fileName) => {
    setIsLoading(true);
    try {
      await handleDelete(containerPath, fileName);
      const files = await getEndpoint(containerPath);
      setFiles(files);
    } catch (error) {
      console.error("Failed to delete file:", error);
    } finally {
      setIsLoading(false);
    }
  };

  const onFileUpload = async (event) => {
    setIsLoading(true);
    try {
      await handleFileUpload(event, containerPath);
      const updatedFiles = await getEndpoint(containerPath);
      setFiles(updatedFiles);
    } catch (error) {
      console.error("Failed to upload file:", error);
    } finally {
      setIsLoading(false);
    }
  };

  const onContainerClear = async () => {
    const confirmClear = window.confirm(
      `Are you sure you want to clear ${containerPath}?`
    );
    if (!confirmClear) {
      return;
    }
    setIsLoading(true);
    try {
      await handleContainerClear(containerPath);
      setFiles([]);
    } catch (error) {
      console.error("Failed to clear container:", error);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Box
      borderWidth="1px"
      borderRadius="lg"
      overflow="hidden"
      p={4}
      bg="white"
      shadow="sm"
    >
      <Input
        type="file"
        id={`${containerPath}-file-upload`}
        style={{ display: "none" }} // hide the default input
        onChange={onFileUpload}
      />
      <Heading size="md" mb={3}>
        {header}{" "}
        {showUploadFileButton && (
          <PlusSquareIcon
            onClick={() =>
              document.getElementById(`${containerPath}-file-upload`).click()
            }
            cursor="pointer" // To show a pointer cursor when hovering over the icon
          />
        )}{" "}
        {/* <DeleteIcon onClick={() => onContainerClear()} cursor="pointer" /> */}
        <RepeatIcon onClick={() => onRefresh()} cursor="pointer" />
      </Heading>
      <List spacing={2}>
        {currentFiles.map((file, idx) => (
          <FileListItem
            key={`${file.Name}-${idx}`}
            file={file}
            onDelete={onDelete}
            onDownload={onDownload}
          />
        ))}
      </List>
      <Stack
        direction="row"
        spacing={4}
        justifyContent="center"
        alignItems="center"
        mt={4}
      >
        <Button
          onClick={() => paginate(1)}
          isDisabled={currentPage === 1}
          leftIcon={<ChevronLeftIcon />}
        >
          First
        </Button>
        <Button
          onClick={() => paginate(currentPage - 1)}
          isDisabled={currentPage === 1}
          leftIcon={<ChevronLeftIcon />}
        >
          Prev
        </Button>
        <Button
          onClick={() => paginate(currentPage + 1)}
          isDisabled={currentPage === totalPages}
          rightIcon={<ChevronRightIcon />}
        >
          Next
        </Button>
        <Button
          onClick={() => paginate(totalPages)}
          isDisabled={currentPage === totalPages}
          rightIcon={<ChevronRightIcon />}
        >
          Last
        </Button>
      </Stack>
    </Box>
  );
};

export default FileTable;
