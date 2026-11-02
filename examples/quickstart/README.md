# Quickstart: Redshift -> PostgreSQL

A minimal, working deployment: [`aurora-exporter.yaml`](./aurora-exporter.yaml)
declares three pipelines, and [`migrations/`](./migrations) creates the
target tables ([golang-migrate](https://github.com/golang-migrate/migrate)
format).

| Pipeline      | Source                         | Target                  | Mode      |
|---------------|--------------------------------|-------------------------|-----------|
| `customers`   | `sales.customers`              | `analytics.customers`   | `replace` |
| `orders`      | `sales.orders`, last 2 days    | `analytics.orders`      | `upsert`  |
| `daily_sales` | `sales.orders`, grouped by day | `analytics.daily_sales` | `replace` |

## Run it locally

`docker-compose.yml` provides a PostgreSQL that stands in for Redshift
(same wire protocol and driver) and a target PostgreSQL. The config's
DSN defaults already point at them.

```bash
make compose-up build

# Demo data in the source, schema in the target.
docker compose exec -T source   psql -q -U test -d source < examples/quickstart/source-seed.sql
docker compose exec -T postgres psql -q -U test -d target < examples/quickstart/migrations/000001_init.up.sql

./bin/aurora-exporter validate -ping -config examples/quickstart/aurora-exporter.yaml
./bin/aurora-exporter run -all       -config examples/quickstart/aurora-exporter.yaml

make compose-down
```

## Against real databases

```bash
export REDSHIFT_DSN='postgres://user:pass@cluster.region.redshift.amazonaws.com:5439/dev?sslmode=require'
export TARGET_DSN='postgres://user:pass@cluster.region.rds.amazonaws.com:5432/app?sslmode=require'

migrate -database "$TARGET_DSN" -path examples/quickstart/migrations up
aurora-exporter validate -ping -config examples/quickstart/aurora-exporter.yaml
aurora-exporter serve -config examples/quickstart/aurora-exporter.yaml
```

To export your own tables, change the source queries, add the target
tables as a new migration (`000002_...`), and keep column names aligned:
rows are inserted by column name, so aliases in the query must match the
target columns.
