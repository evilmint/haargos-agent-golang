#!/usr/bin/with-contenv bashio
set -e

# Configuration
CONFIG_PATH=/data/options.json
declare agent_token
declare debug_mode

agent_token=$(bashio::config 'agent_token')
debug_mode=$(bashio::config.true 'debug_mode')
HA_CONFIG="/config/"

if [ -z "${agent_token}" ]; then
    bashio::log.error "Agent token is not set. Exiting..."
    exit 1
fi

# Start Haargos

bashio::log.info "Starting Haargos..."

if bashio::config.true 'debug_mode'; then
    bashio::log.info "Debug mode is enabled."
fi

DEBUG="${debug_mode}" HAARGOS_AGENT_TOKEN="${agent_token}" ./haargos run --agent-type addon --zha-path "${HA_CONFIG}zigbee.db" --ha-config "${HA_CONFIG}"