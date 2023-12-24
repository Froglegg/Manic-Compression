package serviceBus

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	audioTypes "manic-compression/pkg/audio_types"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
)

// see https://learn.microsoft.com/en-us/azure/service-bus-messaging/service-bus-go-how-to-use-queues

type Msg struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

func (m *Msg) Serialize() string {
	msgBytes, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return string(msgBytes)
}

func (m *Msg) Deserialize(msgBody []byte) {
	err := json.Unmarshal(msgBody, m)
	if err != nil {
		panic(err)
	}
}

type ServiceBus struct {
	client *azservicebus.Client
}

const (
	TaskCompleted  = "Completed"
	TaskInProgress = "In Progress"
)

func NewServiceBus() *ServiceBus {

	connectionString, ok := os.LookupEnv("AZURE_SERVICEBUS_CONNECTION_STRING")
	if !ok {
		panic("AZURE_SERVICEBUS_CONNECTION_STRING environment variable not found")
	}

	client, err := azservicebus.NewClientFromConnectionString(connectionString, nil)
	if err != nil {
		panic(err)
	}

	return &ServiceBus{
		client: client,
	}

}
func (sb *ServiceBus) SendMessage(
	message Msg,
	queue string,
) {

	jsonMessage, err := json.Marshal(message)
	if err != nil {
		panic(err)
	}

	sender, err := sb.client.NewSender(queue, nil)
	if err != nil {
		panic(err)
	}
	defer sender.Close(context.TODO())

	sbMessage := &azservicebus.Message{
		Body: jsonMessage,
	}
	err = sender.SendMessage(context.TODO(), sbMessage, nil)
	if err != nil {
		panic(err)
	}
}

func (sb *ServiceBus) SendMessageBatch(
	messages []Msg,
	queue string,
) {
	sender, err := sb.client.NewSender(queue, nil)
	if err != nil {
		panic(err)
	}
	defer sender.Close(context.TODO())

	batch, err := sender.NewMessageBatch(context.TODO(), nil)
	if err != nil {
		panic(err)
	}

	for _, message := range messages {
		jsonMessage, err := json.Marshal(message)
		if err != nil {
			panic(err)
		}

		if err := batch.AddMessage(&azservicebus.Message{Body: jsonMessage}, nil); err != nil {
			panic(err)
		}
	}
	if err := sender.SendMessageBatch(context.TODO(), batch, nil); err != nil {
		panic(err)
	}
}

func (sb *ServiceBus) GetMessage(count int, queue string) {
	receiver, err := sb.client.NewReceiverForQueue(queue, nil)
	if err != nil {
		panic(err)
	}
	defer receiver.Close(context.TODO())

	messages, err := receiver.ReceiveMessages(context.TODO(), count, nil)
	if err != nil {
		panic(err)
	}

	for _, message := range messages {
		var messageData Msg
		if err := json.Unmarshal(message.Body, &messageData); err != nil {
			fmt.Println("Error unmarshalling message:", err)
			continue
		}

		fmt.Printf("Received message: %+v\n", messageData)

		// CompleteMessage marks the message as complete which removes it from the queue
		err = receiver.CompleteMessage(context.TODO(), message, nil)
		if err != nil {
			panic(err)
		}
	}
}

// for messages that exceed deadlines, or are otherwise invalid, you can dead letter them
func (sb *ServiceBus) DeadLetterMessage(queue string) {
	deadLetterOptions := &azservicebus.DeadLetterOptions{
		ErrorDescription: to.Ptr("exampleErrorDescription"),
		Reason:           to.Ptr("exampleReason"),
	}

	receiver, err := sb.client.NewReceiverForQueue(queue, nil)
	if err != nil {
		panic(err)
	}
	defer receiver.Close(context.TODO())

	messages, err := receiver.ReceiveMessages(context.TODO(), 1, nil)
	if err != nil {
		panic(err)
	}

	if len(messages) == 1 {
		err := receiver.DeadLetterMessage(context.TODO(), messages[0], deadLetterOptions)
		if err != nil {
			panic(err)
		}
	}
}

func (sb *ServiceBus) GetDeadLetterMessage(queue string) {
	receiver, err := sb.client.NewReceiverForQueue(
		queue,
		&azservicebus.ReceiverOptions{
			SubQueue: azservicebus.SubQueueDeadLetter,
		},
	)
	if err != nil {
		panic(err)
	}
	defer receiver.Close(context.TODO())

	messages, err := receiver.ReceiveMessages(context.TODO(), 1, nil)
	if err != nil {
		panic(err)
	}

	for _, message := range messages {
		fmt.Printf("DeadLetter Reason: %s\nDeadLetter Description: %s\n", *message.DeadLetterReason, *message.DeadLetterErrorDescription) //change to struct an unmarshal into it
		err := receiver.CompleteMessage(context.TODO(), message, nil)
		if err != nil {
			panic(err)
		}
	}
}

func (sb *ServiceBus) PeekQueue(queue string) (map[string]audioTypes.AudioTask, error) {
	tasks := make(map[string]audioTypes.AudioTask)

	receiver, err := sb.client.NewReceiverForQueue(queue, nil)
	if err != nil {
		return nil, err
	}
	defer receiver.Close(context.TODO())

	for {
		// peek at the next 10 messages
		messages, err := receiver.PeekMessages(context.Background(), 10, nil)
		if err != nil {
			return nil, err
		}

		for _, message := range messages {
			msg := Msg{}
			task := audioTypes.AudioTask{}
			msg.Deserialize(message.Body)
			task.Deserialize([]byte(msg.Content))
			tasks[task.TaskID] = task
		}

		if len(messages) < 10 {
			// no more messages in the queue
			break
		}
	}

	return tasks, nil
}

func (sb *ServiceBus) ClearQueue(queue string) error {
	receiver, err := sb.client.NewReceiverForQueue(queue, nil)
	if err != nil {
		return err
	}
	defer receiver.Close(context.Background())

	for {
		ctxTimeout, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		messages, err := receiver.ReceiveMessages(ctxTimeout, 10, nil)
		if err != nil {
			if err == context.DeadlineExceeded {
				break // exit if no more messages are received within the timeout
			}
			return err
		}

		for _, message := range messages {
			// complete each message to remove it from the queue
			err := receiver.CompleteMessage(context.Background(), message, nil)
			if err != nil {
				return err
			}
		}

		if len(messages) < 10 {
			// no more messages in the queue
			break
		}
	}

	return nil
}
