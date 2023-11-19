#!/usr/bin/with-contenv bashio
set -e

# Configuration
CONFIG_PATH=/data/options.json

AGENT_TOKEN=$(bashio::config "agent_token")
HA_CONFIG="/config"

# Start Haargos

bashio::log.info "Starting Haargos..."

./haargos run --agent-token "${AGENT_TOKEN}" --ha-config "${HA_CONFIG}"