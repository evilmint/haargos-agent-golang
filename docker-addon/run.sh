#!/bin/sh
set -e

if [ -z "${HAARGOS_AGENT_TOKEN}" ]; then
    echo "Agent token is not set. Exiting..."
    exit 1
fi

# Set STAGE to 'production' if not set
STAGE=${STAGE:-production}

# Start Haargos
echo "Starting Haargos..."

STAGE="${STAGE}" DEBUG="${DEBUG}" HAARGOS_AGENT_TOKEN="${HAARGOS_AGENT_TOKEN}" ./haargos run --agent-type docker --zha-path "/config/zigbee.db" --ha-config "/config/"
