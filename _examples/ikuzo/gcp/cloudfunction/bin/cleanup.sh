#!/usr/bin/env sh

rm -rf vendor

gcloud functions delete ikuzo-simple \
  --region europe-west1
