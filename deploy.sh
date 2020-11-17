#!/bin/bash

# goをビルドし、デプロイする
cd lambda
GOARCH=amd64 GOOS=linux go build -o bin/main
cd ../
cdk deploy