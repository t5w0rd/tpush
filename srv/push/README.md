# Push Service

This is the Push service

Generated with

```
micro new tpush/srv/push --namespace=tpush --type=srv
```

## Getting Started

- [Configuration](#configuration)
- [Dependencies](#dependencies)
- [Usage](#usage)

## Configuration

- FQDN: tpush.srv.push
- Type: srv
- Alias: push

## Dependencies

Micro services depend on service discovery. The default is multicast DNS, a zeroconf system.

In the event you need a resilient multi-host setup we recommend etcd.

```
# install etcd
brew install etcd

# run etcd
etcd
```

## Usage

A Makefile is included for convenience

Build the binary

```
make build
```

Run the service
```
./push-srv
```

Build a docker image
```
make docker
```
