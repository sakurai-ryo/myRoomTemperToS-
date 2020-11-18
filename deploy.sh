#!/bin/bash

# goをビルドし、デプロイする
cd lambda
echo "----------Go Build ----------"
GOARCH=amd64 GOOS=linux go build -o bin/main
echo "Done"

cd ../
echo "----------Deploy----------"
cdk deploy