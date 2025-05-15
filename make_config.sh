#!/usr/bin/env bash

# shellcheck disable=SC1073
cat <<END >"$HOME"/.warp-server.yaml
cisco_host: 'vpn-profile' # Vpn profile name
cisco_username: 'your-cisco-username'
cisco_password: 'your-cisco-password'
local_username: 'your-local-username'
local_password: 'your-local-password'
localhost: '127.0.0.1' # WARN Always this local address
tunnel_address: '192.168.64.8:8080' # WARN UNUSED Virtual Machine IP
daemon_mode: false # Application by default(false) or daemon mode(autorun on start up)
vpn_only: false # Manage only Vpn connection by default disabled(false)

END
