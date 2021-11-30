# opsgenie-exporter

**Prometheus Exporter for Atlassian [Opsgenie](https://www.atlassian.com/de/software/opsgenie).**

## Installation

For pre-built binaries please take a look at the releases.  
https://github.com/cbrgm/opsgenie-exporter/releases

You will need an **Opsgenie API Key** for this exporter to work. Create one via the Opsgenie UI at

* Settings -> API Key Management -> Add new API Key.

The API Key needs the following permissions:

* Read (querying Alerts)
* Configuration Access (querying Teams and Users)

### Container Usage

```bash
docker pull quay.io/cbrgm/opsgenie-exporter:latest
docker run --rm -p 9212:9212 quay.io/cbrgm/opsgenie-exporter --opsgenie.apikey=<id here>
```

## Usage

```bash
Usage: opsgenie-exporter --opsgenie.apikey=STRING

Flags:
  -h, --help                        Show context-sensitive help.
      --http.addr="0.0.0.0:9212"    The address the exporter is running on
      --http.path="/metrics"        The path metrics will be exposed at
      --log.json                    Tell the exporter to log json and not key value pairs
      --log.level="info"            The log level to use for filtering logs
      --opsgenie.apikey=STRING      The opsgenie api token

```

## Metrics

|Name                                         |Type     |Cardinality   |Help
|----                                         |----     |-----------   |----
| opsgenie_alert_count                 | gauge   | 1            | Returns the total amount of alerts. Can be filtered by status `closed`, `open` or all
| opsgenie_team_count                | gauge   | 1            | Returns the number of teams in your account
| opsgenie_user_count                | gauge   | 1            | Returns the number of users. Can be selected by `role`

## Development

```bash
go get -u github.com/cbrgm/opsgenie-exporter
```

## Contributing & License

Feel free to submit changes! See
the [Contributing Guide](https://github.com/cbrgm/contributing/blob/master/CONTRIBUTING.md). This project is open-source
and is developed under the terms of
the [Apache 2.0 License](https://github.com/cbrgm/opsgenie-exporter/blob/master/LICENSE).
