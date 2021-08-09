# Coupling

Strong coupling makes your Web App hard to maintain over time.

The first example shows a tightly coupled service, and the following three a few tactics to approach loose coupling.

All examples cover the same domain and work in the same way.

## Examples

1. [Tightly Coupled](./01-tightly-coupled)
2. [Loosely Coupled](./02-loosely-coupled)
3. [Loosely Coupled with Code Generation](./03-loosely-coupled-generated)
4. [Loosely Coupled with Application Layer](./04-loosely-coupled-app-layer)

## Tests

The [`tests`](./tests) directory holds end-to-end tests for all examples. All applications work the same and expose the same API.

## Running

The [docker-compose definition](./docker-compose.yml) holds all services and their dependencies. Run it with:

```bash
docker-compose up
```

Then, run end-to-end tests for all examples at once:

```bash
make test
```

## Lines of code comparison

### 01-tightly-coupled

Language|files|blank|comment|code
:-------|-------:|-------:|-------:|-------:
Go|4|60|0|292
SQL|1|0|0|1
--------|--------|--------|--------|--------
SUM:|5|60|0|293

### 02-loosely-coupled

Language|files|blank|comment|code
:-------|-------:|-------:|-------:|-------:
Go|3|72|0|345
SQL|1|0|0|1
--------|--------|--------|--------|--------
SUM:|4|72|0|346

### 03-loosely-coupled-generated

Language|files|blank|comment|code
:-------|-------:|-------:|-------:|-------:
Go|11|549|175|2452
YAML|1|7|0|124
SQL|1|2|0|21
TOML|1|0|0|5
--------|--------|--------|--------|--------
SUM:|14|558|175|2602

### 04-loosely-coupled-app-layer

Language|files|blank|comment|code
:-------|-------:|-------:|-------:|-------:
Go|12|575|181|2562
YAML|1|7|0|124
SQL|1|2|0|21
TOML|1|0|0|5
--------|--------|--------|--------|--------
SUM:|15|584|181|2712

