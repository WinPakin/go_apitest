#/bin/bash

protoc ackpb/ack.proto --go_out=plugins=grpc:.
