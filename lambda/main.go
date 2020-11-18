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
	S string `json:"S, string"`
}
type IntNewImage struct {
	N float64 `json:"N, float64"`
}
type StreamEntity struct {
	Timestamp StringNewImage `json:"timestamp"`
	Temper    IntNewImage    `json:"temper"`
}

type Entity struct {
	Timestamp events.DynamoDBAttributeValue `json:"timestamp, string"`
	Temper    events.DynamoDBAttributeValue `json:"temper, float64"`
}

const maxUint = 4294967295

// firehoseにデータを投入する
func firehoseHandler(timestamp events.DynamoDBAttributeValue, temper events.DynamoDBAttributeValue) error {
	fmt.Println("時刻: ", timestamp)
	fmt.Println("気温: ", temper)

	streamName := os.Getenv("FIREHOSE_NAME")

	sess := session.Must(session.NewSession())
	firehoseService := firehose.New(sess, aws.NewConfig().WithRegion("ap-northeast-1"))

	recordsBatchInput := &firehose.PutRecordBatchInput{}
	recordsBatchInput = recordsBatchInput.SetDeliveryStreamName(streamName)

	records := []*firehose.Record{}

	data := Entity{
		Timestamp: timestamp,
		Temper:    temper,
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
	for _, record := range e.Records {

		timestamp := record.Change.NewImage["temper"]
		temper := record.Change.NewImage["timestamp"]

		// s, err := json.Marshal(record.Change.NewImage)
		// if err != nil {
		// 	fmt.Println(err)
		// }
		// fmt.Println("stream =>", string(s))

		// var newImage StreamEntity
		// if err := json.Unmarshal(s, &newImage); err != nil {
		// 	fmt.Println(err)
		// }
		// fmt.Println("newImage: ", newImage)

		if err := firehoseHandler(timestamp, temper); err != nil {
			fmt.Println(err)
		}
	}
}

func main() {
	lambda.Start(Handler)
}
