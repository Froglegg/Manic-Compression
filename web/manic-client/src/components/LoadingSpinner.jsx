import React from "react";
import { Flex, Spinner } from "@chakra-ui/react";

const LoadingSpinner = () => (
  <Flex
    position="fixed"
    top="0"
    right="0"
    bottom="0"
    left="0"
    justifyContent="center"
    alignItems="center"
    background="whiteAlpha.500"
    zIndex="modal"
    pointerEvents="none"
  >
    <Spinner size="xl" />
  </Flex>
);

export default LoadingSpinner;
