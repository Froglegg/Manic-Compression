package audioTypes

import (
	"encoding/json"
)

// audio job specifies the input file and the audio functions to be applied to it, this will be serialized into
// a message and sent to the service bus for processing by Azure Durable Functions
type AudioTask struct {
	ClientID              string   `json:"clientID"`
	TaskID                string   `json:"taskID"`
	Status                string   `json:"status"`
	InputFile             string   `json:"inputFile"`
	OutputFile            string   `json:"outputFile"`
	AudioFunctionPipeline []string `json:"audioFunctionPipeline"`
}

const (
	AudioFunctionWAVToMp3     = "WAV to MP3"
	AudioFunctionApplyEffect1 = "Apply Effect 1"
	AudioFunctionApplyEffect2 = "Apply Effect 2"
)

var audioFunctionMap = map[string]string{
	AudioFunctionWAVToMp3:     "WavToMP3",
	AudioFunctionApplyEffect1: "ApplyEffect1",
	AudioFunctionApplyEffect2: "ApplyEffect2",
}

func (at *AudioTask) Serialize() string {
	// serialize the audio function pipeline to the function names that will be used by the Azure Durable Function
	for idx, titleCaseName := range at.AudioFunctionPipeline {
		if activityName, ok := audioFunctionMap[titleCaseName]; ok {
			at.AudioFunctionPipeline[idx] = activityName
		}
	}
	jobBytes, err := json.Marshal(at)
	if err != nil {
		panic(err) // or handle the error in a way that's appropriate for your application
	}

	return string(jobBytes)
}

func (at *AudioTask) Deserialize(taskBody []byte) {
	err := json.Unmarshal(taskBody, at)
	if err != nil {
		panic(err) // or handle the error in a way that's appropriate for your application
	}
}
