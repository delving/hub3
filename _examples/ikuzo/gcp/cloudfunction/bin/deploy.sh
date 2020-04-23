#!/bin/bash

# update go modules
go mod tidy

# create vendor directory 
go mod vendor

# deploy cloud function
gcloud functions deploy ikuzo-simple \
  --entry-point F \
  --memory 128MB \
  --region europe-west1 \
  --runtime go111 \
  --trigger-http

