import React, { useState, useEffect } from "react";
import {
  GridItem,
  FormControl,
  FormLabel,
  List,
  ListItem,
  Checkbox,
  Box,
  Select,
} from "@chakra-ui/react";
import { functionStrMap } from "../utils.js";

const AudioFunctionSelector = ({
  functions,
  selectedAudioFunctions,
  setSelectedAudioFunctions,
  setAudioFunctionPipeline,
}) => {
  const [functionOrder, setFunctionOrder] = useState({});

  useEffect(() => {
    setAudioFunctionPipeline(
      Object.entries(functionOrder)
        .sort((a, b) => a[1] - b[1])
        .map((entry) => entry[0])
    );
  }, [functionOrder, setAudioFunctionPipeline]);

  const handleCheckboxChange = (functionName) => {
    if (selectedAudioFunctions.includes(functionName)) {
      setSelectedAudioFunctions(
        selectedAudioFunctions.filter((func) => func !== functionName)
      );
      const newOrder = { ...functionOrder };
      delete newOrder[functionName];
      setFunctionOrder(newOrder);
    } else {
      setSelectedAudioFunctions([...selectedAudioFunctions, functionName]);
      setFunctionOrder({
        ...functionOrder,
        [functionName]: Object.keys(functionOrder).length + 1,
      });
    }
  };

  const handleOrderChange = (functionName, order) => {
    if (order === "") {
      // If the placeholder is selected, remove the function from the order
      const newOrder = { ...functionOrder };
      delete newOrder[functionName];
      setFunctionOrder(newOrder);
    } else {
      // Otherwise, update the order as usual
      setFunctionOrder({
        ...functionOrder,
        [functionName]: parseInt(order, 10),
      });
    }
  };

  const getOrderOptions = (currentFunction) => {
    const usedOrders = new Set(Object.values(functionOrder));
    return functions.map((_, index) => {
      const order = index + 1;
      if (!usedOrders.has(order) || functionOrder[currentFunction] === order) {
        return (
          <option key={order} value={order}>
            {order}
          </option>
        );
      }
      return null;
    });
  };

  return (
    <GridItem>
      <FormControl id="map-function">
        <FormLabel>Audio Processing Pipeline</FormLabel>
        <Box border="1px" borderColor="gray.200" borderRadius="md" p={4}>
          <List spacing={2}>
            {functions.map((func, idx) => (
              <ListItem key={`${func}-${idx}`}>
                <Checkbox
                  isChecked={selectedAudioFunctions.includes(func)}
                  onChange={() => handleCheckboxChange(func)}
                >
                  {functionStrMap[func]}
                </Checkbox>
                {selectedAudioFunctions.includes(func) && (
                  <Select
                    placeholder="Select order"
                    value={functionOrder[func] || ""}
                    onChange={(e) => handleOrderChange(func, e.target.value)}
                    ml={2}
                  >
                    {getOrderOptions(func)}
                  </Select>
                )}
              </ListItem>
            ))}
          </List>
        </Box>
      </FormControl>
    </GridItem>
  );
};

export default AudioFunctionSelector;
