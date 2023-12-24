import React, { useEffect, useState } from "react";
import {
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  Button,
  Box,
  Flex,
  Spacer,
} from "@chakra-ui/react";

import {
  getActiveTasks,
  getCompletedTasks,
  clearActiveTasks,
  clearCompletedTasks,
} from "../api"; // Ensure clearAllTasks is defined in your API functions

import { functionStrMap } from "../utils.js";

const POLL_INTERVAL = 1500;
const TaskStatus = ({ queue }) => {
  const [tasks, setTasks] = useState([]);

  useEffect(() => {
    const fetchActiveTasks = async () => {
      let statuses;
      if (queue === "active") {
        statuses = await getActiveTasks();
      } else {
        statuses = await getCompletedTasks();
      }
      if (statuses) {
        setTasks(Object.entries(statuses)); // convert to array of key, value pairs
      }
    };
    const intervalId = setInterval(fetchActiveTasks, POLL_INTERVAL);
    return () => clearInterval(intervalId);
  }, [queue]);

  const handleClearAllTasks = async () => {
    if (queue === "active") {
      await clearActiveTasks();
    } else {
      await clearCompletedTasks();
    }
    setTasks([]);
  };

  return (
    <Box>
      <Flex mb={4} justifyContent="flex-start">
        <Spacer />
        <Button
          aria-label="Clear All Tasks"
          onClick={handleClearAllTasks}
          colorScheme="red"
        >
          Clear Tasks
        </Button>
      </Flex>
      <Table variant="simple">
        <Thead>
          <Tr>
            <Th>Task ID</Th>
            <Th>Input File</Th>
            <Th>Status</Th>
            <Th>Processing Pipeline</Th>
          </Tr>
        </Thead>
        <Tbody>
          {tasks.map(([taskID, task]) => (
            <Tr key={taskID}>
              <Td>{taskID}</Td>
              <Td>{task.inputFile}</Td>
              <Td>{task.status}</Td>
              <Td>
                {task.audioFunctionPipeline
                  .map((func) => functionStrMap[func])
                  .join(", ")}
              </Td>
            </Tr>
          ))}
        </Tbody>
      </Table>
    </Box>
  );
};

export default TaskStatus;
