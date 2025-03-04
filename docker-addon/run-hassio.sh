#!/usr/bin/with-contenv bashio
set -e

# Configuration
CONFIG_PATH=/data/options.json
declare agent_token
declare ha_access_token

agent_token=$(bashio::config 'agent_token')
ha_access_token=$(bashio::config 'ha_access_token')
HA_CONFIG="/config/"

if [ -z "${agent_token}" ]; then
    bashio::log.error "Agent token is not set. Exiting..."
    exit 1
fi

# Start Haargos

bashio::log.info "Starting Haargos..."

XSTAGE="production"
if bashio::config.true 'dev'; then
    XSTAGE="dev"
    bashio::log.info "Connecting to dev."
fi

XDEBUG="false"
# Check for debug mode without causing the script to exit if debug mode is false
if bashio::config.true 'debug_mode'; then
    XDEBUG="true"
    bashio::log.info "Debug mode is enabled."
fi

STAGE="${XSTAGE}" DEBUG="${XDEBUG}" HA_ACCESS_TOKEN="${ha_access_token}" HAARGOS_AGENT_TOKEN="${agent_token}" ./haargos run --agent-type addon --zha-path "${HA_CONFIG}zigbee.db" --ha-config "${HA_CONFIG}"
