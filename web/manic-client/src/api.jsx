import axios from "axios";
// if deployed as build script in cluster, nginx handles proxy to /api
// else, if running locally, react handles proxy to /api
const API_PATH = "/api";

const getEndpoint = async (endpoint) => {
  const res = await axios.get(API_PATH + endpoint);
  return res.data;
};

const handleFileUpload = async (event, endpoint) => {
  const uploadedFiles = event.target.files;
  const formData = new FormData();
  for (const file of uploadedFiles) {
    formData.append("files", file);
  }
  const res = await axios.post(API_PATH + endpoint, formData);
  return res.data;
};

const handleFileDownload = async (containerPath, fileName) => {
  try {
    const response = await fetch(API_PATH + `${containerPath}${fileName}`);
    const blob = await response.blob();

    const link = document.createElement("a");
    link.href = window.URL.createObjectURL(blob);
    link.download = fileName;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    window.URL.revokeObjectURL(link.href);
  } catch (error) {
    console.error("Download failed:", error.message);
  }
};

const handleStart = async (job) => {
  const res = await axios.post(API_PATH + "/start", {
    inputFiles: job.inputFiles,
    clientID: job.clientID,
    audioFunctionPipeline: job.audioFunctionPipeline,
  });
  return res.data;
};

const getActiveTasks = async () => {
  try {
    const response = await axios.get(`${API_PATH}/activeTasks`);
    return response.data;
  } catch (error) {
    console.error("Error polling task status:", error);
    return null;
  }
};

const getCompletedTasks = async () => {
  try {
    const response = await axios.get(`${API_PATH}/completedTasks`);
    return response.data;
  } catch (error) {
    console.error("Error polling task status:", error);
    return null;
  }
};

const clearActiveTasks = async () => {
  try {
    const response = await axios.post(`${API_PATH}/clearActiveTasks`);
    console.log(response.data);
  } catch (error) {
    console.error("Error clearing tasks:", error);
  }
};

const clearCompletedTasks = async () => {
  try {
    const response = await axios.post(`${API_PATH}/clearCompletedTasks`);
    console.log(response.data);
  } catch (error) {
    console.error("Error clearing tasks:", error);
  }
};

const handleDelete = async (containerPath, fileName) => {
  const filePath = `${containerPath}${fileName}`;
  const res = await axios.delete(API_PATH + filePath);
  return res.data;
};

const handleContainerClear = async (containerPath) => {
  const res = await axios.delete(API_PATH + containerPath);
  return res.data;
};

export {
  handleFileUpload,
  handleFileDownload,
  handleStart,
  handleDelete,
  getEndpoint,
  handleContainerClear,
  getActiveTasks,
  clearActiveTasks,
  clearCompletedTasks,
  getCompletedTasks,
};
