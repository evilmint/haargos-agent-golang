#!/usr/bin/with-contenv bashio
set -e

# Configuration
CONFIG_PATH=/data/options.json
declare agent_token

agent_token=$(bashio::config 'agent_token')
HA_CONFIG="/config/"

if [ -z "${agent_token}" ]; then
    bashio::log.error "Agent token is not set. Exiting..."
    exit 1
fi

# Start Haargos

bashio::log.info "Starting Haargos..."

./haargos run --agent-token "${agent_token}" --ha-config "${HA_CONFIG}"