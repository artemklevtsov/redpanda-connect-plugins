# Redpanda Connect for Yandex.Metrika API

Redpanda connect plugins to fetch data from the Yandex.Metrika API.

## Build

Finally, build your custom main func:

```sh
go install github.com/artemklevtsov/redpanda-connect-yandex-metrika@latest
```

Alternatively build it as a Docker image with:

```sh
git clone https://github.com/artemklevtsov/redpanda-connect-yandex-metrika
cd redpanda-connect-yandex-metrika
docker build -t redpanda-connect-yandex-metrika .
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
