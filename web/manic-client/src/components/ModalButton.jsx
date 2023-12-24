import React from "react";
import {
  Box,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalCloseButton,
  Button,
  useDisclosure,
} from "@chakra-ui/react";

const ModalButton = ({
  modalButtonText,
  modalHeaderText,
  children,
  size = "lg",
  mx = 0,
}) => {
  const { isOpen, onOpen, onClose } = useDisclosure();

  return (
    <Box position="relative" zIndex="docked">
      <Button onClick={onOpen} mx={mx}>
        {modalButtonText}
      </Button>

      <Modal isOpen={isOpen} onClose={onClose} size={size}>
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>{modalHeaderText}</ModalHeader>
          <ModalCloseButton />
          <ModalBody overflowX="auto">{children}</ModalBody>
        </ModalContent>
      </Modal>
    </Box>
  );
};

export default ModalButton;
