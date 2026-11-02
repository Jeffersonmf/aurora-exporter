# Example: Redshift -> Aurora PostgreSQL ("hotdata")

The original deployment this project was built for, expressed as
configuration: [`aurora-exporter.yaml`](./aurora-exporter.yaml) holds
the 21 exports that used to be 21 Go files, and
[`migrations/`](./migrations) creates their target tables
([golang-migrate](https://github.com/golang-migrate/migrate) format).

```bash
export REDSHIFT_DSN='postgres://user:pass@datalake.example.redshift.amazonaws.com:5439/datahub?sslmode=require'
export HOTDATA_DSN='postgres://user:pass@hotdata.cluster.example.rds.amazonaws.com:5432/hotdata?sslmode=require'

migrate -database "$HOTDATA_DSN" -path examples/redshift-to-aurora/migrations up
aurora-exporter validate -config examples/redshift-to-aurora/aurora-exporter.yaml -ping
aurora-exporter run -all -config examples/redshift-to-aurora/aurora-exporter.yaml
```

Note: migration `000004` has a typo in its up file name
(`colab_uber`) that was kept as-is, since golang-migrate identifies
migrations by number and renaming it would not change behavior.
