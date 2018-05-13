# ShadowSocks Monitoring Service

Read multiple shadowsocks server config files, and start local clients using `ss-local`.

The monitor will HTTP ping google every 10 seconds, and restart a ss-server if the ping fails.

A sample config file for multiple ss servers:

```
{
  "hk2": {
    "server": "hk2.example.net",
    "server_port": 2333,
    "local_address": "127.0.0.1",
    "password": "***",
    "timeout": 600,
    "method": "chacha20",
    "fast_open": true
  },
  "hk2": {
    "server": "hk2.example.net",
    "server_port": 2333,
    "local_address": "127.0.0.1",
    "password": "***",
    "timeout": 600,
    "method": "chacha20",
    "fast_open": true
  },
  "hk6": {
    "server": "hk.example.net",
    "server_port": 2333,
    "local_address": "127.0.0.1",
    "password": "***",
    "timeout": 600,
    "method": "chacha20",
    "fast_open": true
  }
}
```
