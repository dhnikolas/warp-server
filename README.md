# warp-server

![logo](assets/logo.png)

## Install warp-server

```ssh
brew install dhnikolas/tools/warp-server
```

### Configuration

#### Create file ~/.warp-server.yaml

```yaml
cisco_host: 'profile_name' #Vpn profile name
cisco_username: 'john.doe'
cisco_password: '*****'
local_username: 'john'
local_password: '****'
localhost: '127.0.0.1' #Always this local address
tunnel_address: '192.168.64.8:8080' #Virtual Machine IP
daemon_mode: false # Application by default(false) or daemon mode(autorun on start up)
vpn_only: false # Manage only Vpn connection by default disabled(false)
```
