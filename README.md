# Firefox send tools

## Description

Tools for Firefox Send (https://github.com/timvisee/ffsend).

- `send-cleanup`: remove keys in Redis without a corresponding file in GCS
- `send-exporter`: expose storage metrics to Prometheus (TODO)

The former is intended to be used as a `CronJob`, while the latter is an
exported intended to be run as a `Deployment`.

## Installation

TBD

## Usage

TBD

## TODO

Allow other backend storages and use proper backend according to URL, e.g:

- `gs://` -> GCS
- `s3://` -> S3
- `file://` or `^/` -> local file system

## Contributing

TBD

## License

MIT