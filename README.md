# Redpanda Connect for Yandex.Metrika API

Redpanda connect plugins to fetch data from the Yandex.Metrika API.

## Install

Build and Install binary:

```sh
go install github.com/artemklevtsov/redpanda-connect-yandex-metrika@latest
```

Or download binary with:

```sh
curl -s https://i.jpillora.com/artemklevtsov/redpanda-connect-yandex-metrika@latest! | bash
```

Remove `!` if you want to install to the current directory.

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

output:
  stdout: {}
```

And you can run it like this:

```sh
redpanda-connect-yandex-metrika run ./connect.yaml
```

See also `configs/` for the more examples.

## See also

### Authorization

- [debug token](https://yandex.ru/dev/id/doc/en/tokens/debug-token)

### Reporting API

- [dimensions and metrics](https://yandex.ru/dev/metrika/en/stat/attrandmetr/dim_all)

### Logs API

- [hits](https://yandex.ru/dev/metrika/en/logs/fields/hits)
- [visits](https://yandex.ru/dev/metrika/en/logs/fields/visits)
- [parametrization](https://yandex.ru/dev/metrika/en/logs/param)
