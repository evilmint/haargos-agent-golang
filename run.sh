#!/usr/bin/with-contenv bashio
set -e

# Configuration
CONFIG_PATH=/data/options.json

declare agent_token

## Get the 'message' key from the user config options.
agent_token=$(bashio::config 'agent_token')

## Print the message the user supplied, defaults to "Hello World..."
bashio::log.info "${agent_token:="NO_TOKEN"}"

HA_CONFIG="/config"

# Start Haargos

bashio::log.info "Starting Haargos..."

./haargos run --agent-token "${agent_token}" --ha-config "${HA_CONFIG}"