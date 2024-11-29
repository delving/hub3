#!/bin/bash
set -e

# Source the environment script to set the required environment variables
source /fuseki/fuseki-env.sh

# Start the Fuseki server in the foreground
exec /jena-fuseki/fuseki-server --config=/fuseki/template-config.ttl
