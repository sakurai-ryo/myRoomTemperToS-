import * as cdk from "@aws-cdk/core";
import { Table, AttributeType, StreamViewType } from "@aws-cdk/aws-dynamodb";
import {
  AssetCode,
  Function,
  Runtime,
  StartingPosition,
} from "@aws-cdk/aws-lambda";
import { DynamoEventSource } from "@aws-cdk/aws-lambda-event-sources";
import { Bucket } from "@aws-cdk/aws-s3";
import * as firehose from "@aws-cdk/aws-kinesisfirehose";
import {
  Role,
  PolicyStatement,
  ServicePrincipal,
  Effect,
  PolicyDocument,
} from "@aws-cdk/aws-iam";
import * as path from "path";

export class NodeKinesisStack extends cdk.Stack {
  constructor(scope: cdk.Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    // temperを保存するバケット
    const temperBucket = new Bucket(this, "MyEncryptedBucket", {
      bucketName: "temper-bucket",
    });

    // Kinesis Firehose用のポリシー
    const kinesisStatement = new PolicyStatement({
      resources: ["*"],
      effect: Effect.ALLOW,
      actions: [
        "s3:AbortMultipartUpload",
        "s3:GetBucketLocation",
        "s3:GetObject",
        "s3:ListBucket",
        "s3:ListBucketMultipartUploads",
        "s3:PutObject",
      ],
    });

    const kinesisRole = new Role(this, "kinesisFirehoseTemperRole", {
      assumedBy: new ServicePrincipal("firehose.amazonaws.com"),
      roleName: "kinesisFirehoseTemperRole",
      inlinePolicies: {
        firehoseS3Policy: new PolicyDocument({
          statements: [kinesisStatement],
        }),
      },
    });

    // Kinesis Firehose
    const temperFirehose = new firehose.CfnDeliveryStream(
      this,
      "temperFirehose",
      {
        deliveryStreamName: "temperToS3",
        deliveryStreamType: "DirectPut",
        extendedS3DestinationConfiguration: {
          bucketArn: temperBucket.bucketArn,
          roleArn: kinesisRole.roleArn,
        },
      }
    );

    // DynamoDB
    const temperTable = new Table(this, "RoomTemper", {
      partitionKey: {
        name: "timestamp",
        type: AttributeType.STRING,
      },
      sortKey: {
        name: "temper",
        type: AttributeType.NUMBER,
      },
      tableName: "RoomTemper",
      removalPolicy: cdk.RemovalPolicy.DESTROY,
      stream: StreamViewType.NEW_AND_OLD_IMAGES,
    });

    // Lambda
    const temperToS3Lambda = new Function(this, "temperToS3Lambda", {
      code: new AssetCode("lambda/bin"),
      handler: "main",
      runtime: Runtime.GO_1_X,
      environment: {
        TABLE_NAME: temperTable.tableName,
        FIREHOSE_NAME: temperFirehose.deliveryStreamName!,
      },
    });

    // StreamとLambdaの紐付け
    temperToS3Lambda.addEventSource(
      new DynamoEventSource(temperTable, {
        startingPosition: StartingPosition.TRIM_HORIZON,
        batchSize: 5,
        bisectBatchOnError: true,
        retryAttempts: 0,
      })
    );

    // Lambdaにポリシーを推定
    temperToS3Lambda.addToRolePolicy(
      new PolicyStatement({
        resources: ["*"],
        actions: [
          "dynamodb:BatchGetItem",
          "dynamodb:Describe*",
          "dynamodb:List*",
          "dynamodb:GetItem",
          "dynamodb:Query",
          "dynamodb:Scan",
          "firehose:*",
        ],
      })
    );
  }
}
