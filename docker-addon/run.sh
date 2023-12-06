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

echo "ls /"
ls /
echo "ls /data"
ls /data
cat /data/options.json

echo "Agent token is ${agent_token}"
bashio::log.info "Starting Haargos..."

XDEBUG="false"
# Check for debug mode without causing the script to exit if debug mode is false
if bashio::config.true 'debug_mode'; then
    XDEBUG="true"
    bashio::log.info "Debug mode is enabled."
fi

DEBUG="${XDEBUG}" HAARGOS_AGENT_TOKEN="${agent_token}" ./haargos run --agent-type addon --zha-path "${HA_CONFIG}zigbee.db" --ha-config "${HA_CONFIG}"
