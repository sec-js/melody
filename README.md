<p align="center">
  <h1 align="center">Melody</h1>
  <p align="center">Monitor the Internet's background noise</p>
</p>

<p align="center">
    <a href="https://goreportcard.com/badge/github.com/bonjourmalware/melody"><img src="https://goreportcard.com/badge/github.com/bonjourmalware/melody" alt="Go Report Card"/></a>
    <a href="https://coveralls.io/github/bonjourmalware/melody"><img src="https://coveralls.io/repos/github/bonjourmalware/melody/badge.svg" alt="Coverage Status"/></a>
    <a href="https://hub.docker.com/r/bonjourmalware/melody/builds"><img src="https://img.shields.io/docker/cloud/build/bonjourmalware/melody" alt="Docker build status"/></a>
    <a href="https://hub.docker.com/r/bonjourmalware/melody/builds"><img src="https://img.shields.io/docker/image-size/bonjourmalware/melody?sort=date" alt="Docker image size"/></a>
</p>

<p align="center">
       <a href="https://github.com/bonjourmalware/melody/releases/latest"><img src="https://img.shields.io/github/release/bonjourmalware/melody.svg" alt="Latest release"/></a>
    <a href="https://bonjourmalware.github.io/melody/"><img src="https://img.shields.io/badge/%F0%9F%93%9A-Documentation-informational" alt="Documentation"/></a>
    <a href="https://bonjourmalware.github.io/melody/installation"><img src="https://img.shields.io/badge/%F0%9F%93%9A-Installation-informational" alt="Installation"/></a>
    <a href="https://bonjourmalware.github.io/melody/quickstart"><img src="https://img.shields.io/badge/%F0%9F%93%9A-Quickstart-informational" alt="Quickstart"/></a>
    <a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-green.svg" alt="Go Report Card"/></a>
</p>

---

Melody is a transparent internet sensor built for threat intelligence and supported by a detection rule framework which allows you to tag packets of interest for further analysis and threat monitoring.

# Features
Here are some key features of Melody :

+ Transparent capture
+ Write detection rules and tag specific packets to analyze them at scale 
+ Mock vulnerable websites using the builtin HTTP/S server
+ Supports the main internet protocols over IPv4 and IPv6
+ Handles log rotation for you : Melody is designed to run forever on the smallest VPS
+ Minimal configuration required
+ Standalone mode : configure Melody using only the CLI
+ Easily scalable :
    + Statically compiled binary
    + Up-to-date Docker image

Additional features on the roadmap include :

+ Dedicated helper program to create, test and manage rules
+ Centralized rules management

# Use cases

+ Extract trends and patterns from Internet's noise 
+ Index malicious activity, exploitation attempts and targeted scanners
+ Monitor emerging threats exploitation
+ Keep an eye on specific threats
+ Log every packet sent to your application from the internet and block malicious activity

# Quickstart
[Quickstart details.](https://bonjourmalware.github.io/melody/installation)

## TL;DR
### Release
Get the latest release at `https://github.com/bonjourmalware/melody/releases`.

```bash
make install            # Set default outfacing interface
make certs              # Make self signed certs for the HTTPS fileserver
make default_rules      # Enable the default rules
make service            # Create a systemd service to restart the program automatically and launch it at startup 
                        # Note that the script expects that you've installed Melody in /opt/melody

sudo systemctl stop melody  # Stop the service while we're configuring it
```

Update the `filter.bpf` file to filter out unwanted packets.

```bash
sudo systemctl start melody     # Start Melody
sudo systemctl status melody    # Check that Melody is running    
```

The logs should start to pile up in `/opt/melody/logs/melody.ndjson`.

```bash
tail -f /opt/melody/logs/melody.ndjson # | jq
```

### From source

```bash
git clone https://github.com/bonjourmalware/melody /opt/melody
cd /opt/melody
make build
```

Then continue with the steps from the [release](#release) TL;DR.

### Docker

```bash
mkdir -p /opt/melody/logs
cd /opt/melody/

docker pull bonjourmalware/melody:latest

MELODY_CLI="" # Put your CLI options here. Example : MELODY_CLI="-s -o 'http.server.port: 5555'"

docker run \
    --net=host \
    -e "MELODY_CLI=$MELODY_CLI" \
    --mount type=bind,source="$(pwd)"/filter.bpf,target=/app/filter.bpf,readonly \  # Remove this line if you're using the default filter
    --mount type=bind,source="$(pwd)"/config.yml,target=/app/config.yml,readonly \  # Remove this line if you're using the default config
    --mount type=bind,source="$(pwd)"/logs,target=/app/logs/ \                      # The directory must exists in your current directory before running the container
    melody
```

The logs should start to pile up in `/opt/melody/logs/melody.ndjson`.

# Rules

[Rule syntax details.](https://bonjourmalware.github.io/melody/installation)

## Rule example

```yaml
CVE-2020-14882 Oracle Weblogic Server RCE:
  layer: http
  version: 1.0
  meta:
    id: 3e1d86d8-fba6-4e15-8c74-941c3375fd3e
    version: 1.0
    author: BonjourMalware
    status: stable
    created: 2020/11/07
    modified: 2020/11/07
    description: "Checking or trying to exploit CVE-2020-14882"
    references:
      - "https://nvd.nist.gov/vuln/detail/CVE-2020-14882"
  match:
    http.method:
      is:
        - "POST"
    http.uri:
      contains|any:
        - "/console/images/%252E%252E%252Fconsole.portal"
        - "/console/css/%2e"
  tags:
    cve: "cve-2020-14882"
    vendor: "oracle"
    product: "weblogic"
    impact: "rce"
```

# Logs
Example of a line generated by a netcat TCP packet over IPv4 :

```json
{
  "tcp": {
    "window": 512,
    "seq": 1906765553,
    "ack": 2514263732,
    "data_offset": 8,
    "flags": "PA",
    "urgent": 0,
    "payload": {
      "content": "I made a discovery today. I found a computer.\n",
      "base64": "SSBtYWRlIGEgZGlzY292ZXJ5IHRvZGF5LiAgSSBmb3VuZCBhIGNvbXB1dGVyLgo=",
      "truncated": false
    }
  },
  "ip": {
    "version": 4,
    "ihl": 5,
    "tos": 0,
    "length": 99,
    "id": 39114,
    "fragbits": "DF",
    "frag_offset": 0,
    "ttl": 64,
    "protocol": 6
  },
  "timestamp": "2020-11-16T15:50:01.277828+01:00",
  "session": "bup9368o4skolf20rt8g",
  "type": "tcp",
  "src_ip": "127.0.0.1",
  "dst_port": 1234,
  "matches": {},
  "inline_matches": [],
  "embedded": {}
}
```

[Log data details.](https://bonjourmalware.github.io/melody/layers)
