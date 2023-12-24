import React from "react";
import { ListItem, ButtonGroup, IconButton, Text } from "@chakra-ui/react";
import { DeleteIcon, DownloadIcon } from "@chakra-ui/icons";

const FileListItem = ({ file, onDownload, onDelete }) => {
  return (
    <ListItem
      display="flex"
      justifyContent="space-between"
      alignItems="center"
      p={2}
      borderWidth="1px"
      borderRadius="lg"
      mb={2}
    >
      <Text flex="1">
        {file.Name} ({file.Size} bytes)
      </Text>
      <ButtonGroup isAttached variant="outline">
        <IconButton
          aria-label="Download file"
          icon={<DownloadIcon />}
          onClick={() => onDownload(file.Name)}
        />
        <IconButton
          aria-label="Delete file"
          icon={<DeleteIcon />}
          onClick={() => onDelete(file.Name)}
          colorScheme="red"
        />
      </ButtonGroup>
    </ListItem>
  );
};

export default FileListItem;
