#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from '@aws-cdk/core';
import { NodeKinesisStack } from '../lib/node_kinesis-stack';

const app = new cdk.App();
new NodeKinesisStack(app, 'NodeKinesisStack');
