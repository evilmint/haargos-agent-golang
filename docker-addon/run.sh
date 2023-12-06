#!/bin/sh
set -e

if [ -z "${HAARGOS_AGENT_TOKEN}" ]; then
    echo "Agent token is not set. Exiting..."
    exit 1
fi

# Start Haargos

echo "Starting Haargos..."

DEBUG="${DEBUG}" HAARGOS_AGENT_TOKEN="${HAARGOS_AGENT_TOKEN}" ./haargos run --agent-type addon --zha-path "/config/zigbee.db" --ha-config "/config"
