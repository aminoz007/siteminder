# New Relic Siteminder Agent

[![Go Report Card](https://goreportcard.com/badge/github.com/aminoz007/siteminder?style=flat-square)](https://goreportcard.com/report/github.com/aminoz007/siteminder)
[![GoDoc](https://godoc.org/github.com/aminoz007/siteminder?status.svg)](https://godoc.org/github.com/aminoz007/siteminder)
[![Release](https://img.shields.io/github/release/aminoz007/siteminder.svg?style=flat-square)](https://github.com/aminoz007/siteminder/releases/latest)

This is a solution to monitor CA SSO formerly known as CA siteminder. APM EPAgents listens for metrics from SSO clients (web agents or policy servers) on a specific port for metrics data. The port can be configured using the property `introscope.epagent.config.networkDataPort` in `IntroscopeEPAgent.properties` configuration file.
This agent will replace the CA EPAgent so you have to stop your agent first if running. The New Relic siteminder agent will create a server and listens to the `networkDataPort`, once the data is received, the sitemninder agent will collect it, parse it, format it and send it to New Relic.

## Support
- Linux
- Mac
- Windows

## Setup

1. Download the latest release from [here](https://github.com/aminoz007/siteminder/releases).
2. Get your New Relic Insights Insert key: https://insights.newrelic.com/accounts/<ACCOUNT_ID>/manage/api_keys
3. Set up configuration: see [below](https://github.com/aminoz007/siteminder#configuration) for more information

### Configuration
- Two options are provided to configure the siteminder agent: via `siteminder.yml` file or via environement variables. 
- Please note that the environement variables configuration will always overwrite the yaml file configuration
    - Priority: Environement variables -> Yaml file configuration

#### Properties:

| Property (yaml) | Property (env variable) | Required | Default Value | Description
| --- | --- | --- | --- | ---
| insights_url | NR_INSIGHTS_URL | Yes | Not applicable | Insights ingest URL, which should be in this format: https://insights-collector.newrelic.com/v1/accounts/\<yourAccountID\>/events
| insights_key | NR_INSIGHTS_KEY | Yes | Not applicable | Your Insights api insert key.
| port | NR_PORT | Yes | Not applicable |  The port where the data is pushed from siteminder (check the description above for more details).
| host | NR_HOST | No | localhost | The server where the networkDataPort is open
| interval | NR_INTERVAL | No | 30s | FLush interval:  valid time units are **"ns", "us" (or "µs"), "ms", "s", "m", "h"**. This is the time we wait before sending the data to NR.
| custom_attributes | NR_CUSTOM_ATTRS | No | Not applicable | Key value pairs tags used to decorate NR events data. 
| debug | NR_DEBUG | No | false | Verbose logging, useful for debugging.
| max_buffer_size | NR_MAX_BUFFER_SIZE | No | 100 | the maximum buffer size in **KB** (plz, never exceed 1000 which is the POST limit in insights!! otherwise an error will be returned). Data will be sent if interval OR max buffer size is reached whatever comes first.
| max_request_retries | NR_MAX_REQUEST_RETRIES | No | 5 | The maximum number of retries for sending the data when there are network failures.
| proxy_url | NR_PROXY_URL | No | Not applicable | Add your proxy endpoint if you need to send the data to NR via a proxy.


#### Example configuration file via yaml file:
```
insights_url: "https://insights-collector.newrelic.com/v1/accounts/<your account id>/events" # Required; Insights ingest URL
insights_key: "Your Key" # Required; Your insights api insert key
port: "APM for SSO port" # Required; the port where the data is pushed from siteminder
host: "localhost" # Optional; Default localhost, the server where rhe networkDataPort is open
interval: 30s # Optional; Default 30s, FLush interval: valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h";this is the time we wait before sending the data to NR
custom_attributes: # Optional; Key value pairs tags to decorate NR events data.
  owner: nr.expert.services
debug: false # Optional; Default false, useful for debugging
max_buffer_size: 100 # Optional; Default 100KB, the maximum buffer size in KB (plz, never exceed 1000KB which is the POST limit in insights!! otherwise an error will be returned). Data will be sent if interval OR max buffer size is reached whatever comes first
max_request_retries: 5 # Optional; Default 5, the maximum number of retries for sending the data when there are network failures
proxy_url: "Your proxy endpoint" # Optional; add your proxy endpoint if you need to send the data to NR via a proxy
```

## Development
```
Requirements:
- golang installed in your machine

# Setup - Downloads required dependencies
make setup

# Building - Builds required binary as specified below
make build (builds for current OS)
make build-linux (builds for Linux)
make build-darwin (builds for Darwin)
make build-windows (builds for Windows)
make build-all (builds for Linux, Darwin and Windows)

# Running
make run (runs the agent locally)

# Testing
make test (tests locally)

# Packaging - Packages into a tarball including README and config file
make package-linux (packages for Linux)
make package-darwin (packages for Darwin)
make package-windows (packages for Windows)
make package-all (packages for Linux, Darwin and Windows)

# Cleaning up 
make clean (Removes vendor, bin and coverage dirs)

# Other
make version - Outputs agent version
```

## Issues / Enhancement Requests

Issues and enhancement requests can be submitted in the [issues tab of this repository](https://github.com/aminoz007/siteminder/issues).
Please search for and review the existing open issues before submitting a new issue.

## Contributing

Contributions are welcome (and if you submit a Enhancement Request, expect to be invited to
contribute it yourself :grin:).