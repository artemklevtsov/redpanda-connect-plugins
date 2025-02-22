# Redpanda Connect for Yandex.Metrika API

Redpanda connect plugins to fetch data from the Yandex.Metrika API.

## Build

Build and install binary:

```sh
go install github.com/artemklevtsov/redpanda-connect-yandex-metrika@latest
```

Or download binary with:

```sh
curl -s https://i.jpillora.com/artemklevtsov/redpanda-connect-yandex-metrika@latest! | bash
```

Alternatively pull a Docker image with:

```sh
docker pull ghcr.io/artemklevtsov/redpanda-connect-yandex-metrika@latest
```

## Run

```yaml
input:
  yandex_metrika_stat_table:
    ids:
      - 44147844
    metrics:
      - ym:s:users
      - ym:s:visits
    dimensions:
      - ym:s:date
      - ym:s:lastTrafficSource
    sort:
      - ym:s:date
      - ym:s:lastTrafficSource
    date1: 2025-02-01
    date2: 2025-02-28
    # filters: ym:s:lastTrafficSource=='direct'
    # filters: ym:s:date=='2025-02-03'
    format_keys: true

output:
  stdout: {}
```

And you can run it like this:

```sh
redpanda-connect-yandex-metrika run ./connect.yaml
```

See also `configs/` for the more examples.
