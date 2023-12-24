import React from "react";
import { Box, Heading, Container } from "@chakra-ui/react";

const Banner = ({ heading }) => {
  return (
    <Box bg="teal.500" color="white" py={4} mb={4}>
      <Container maxW="container.lg">
        <Heading size="lg">{heading}</Heading>
      </Container>
    </Box>
  );
};

export default Banner;
