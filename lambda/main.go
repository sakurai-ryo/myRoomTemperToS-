package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/firehose"
)

type StringNewImage struct {
	S string
}
type IntNewImage struct {
	N float64
}
type StreamEntity struct {
	timestamp StringNewImage
	temper    IntNewImage
}

type Entity struct {
	timestamp string  `json:"timestamp"`
	temper    float64 `json:"temper"`
}

const maxUint = 4294967295

// firehoseにデータを投入する
func firehoseHandler(timestamp string, temper float64) error {

	streamName := os.Getenv("FIREHOSE_NAME")

	sess := session.Must(session.NewSession())
	firehoseService := firehose.New(sess, aws.NewConfig().WithRegion("ap-northeast-1"))

	recordsBatchInput := &firehose.PutRecordBatchInput{}
	recordsBatchInput = recordsBatchInput.SetDeliveryStreamName(streamName)

	records := []*firehose.Record{}

	data := Entity{
		timestamp: timestamp,
		temper:    temper,
	}

	b, err := json.Marshal(data)

	if err != nil {
		return fmt.Errorf("context: %v", err)
	}

	record := &firehose.Record{Data: b}
	records = append(records, record)

	recordsBatchInput = recordsBatchInput.SetRecords(records)

	resp, err := firehoseService.PutRecordBatch(recordsBatchInput)
	if err != nil {
		return fmt.Errorf("context: %v", err)
	} else {
		fmt.Printf("PutRecordBatch: %v\n", resp)
		return nil
	}
	return nil
}

// Handler lambda
func Handler(ctx context.Context, e events.DynamoDBEvent) {
	fmt.Println(e)

	for _, record := range e.Records {
		s, _ := json.Marshal(record.Change.NewImage)
		fmt.Println("stream =>", string(s))

		var newImage StreamEntity
		json.Unmarshal(s, &newImage)

		err := firehoseHandler(newImage.timestamp.S, newImage.temper.N)

		if err != nil {
			fmt.Println(err)
		}
	}
}

func main() {
	lambda.Start(Handler)
}
