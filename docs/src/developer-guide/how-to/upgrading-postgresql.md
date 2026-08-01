# Upgrading the major version of the local PostgreSQL server
This guide shows the simplest way to upgrade the major version of a
PostgreSQL server that runs as a Docker container. Follow these steps:

1. Start the server
2. Dump the database to a local file using `pg_dump`
3. Stop the server. Delete the Docker volume that holds its data
4. Bump the PostgreSQL version
5. Start the server
6. Restore the database from the local dump file using `pg_restore`


## Start the PostgreSQL database server
```shell
$ docker compose up -d postgres

[+] Running 3/3
 ✔ Network sparklemuffin_default         Created
 ✔ Volume "sparklemuffin_postgres-data"  Created
 ✔ Container sparklemuffin-postgres-1    Started
```

## Dump the database
```shell
$ make pgdump

# mkdir -p dump
# docker compose exec postgres pg_dump -U sparklemuffin sparklemuffin --format custom --compress zstd > dump/sparklemuffin.sql.zst
```

## Stop the PostgreSQL server and delete its Docker volume
```shell
$ docker compose down -v

[+] Running 3/3
 ✔ Container sparklemuffin-postgres-1  Removed
 ✔ Volume sparklemuffin_postgres-data  Removed
 ✔ Network sparklemuffin_default       Removed
```

## Update the PostgreSQL server version
Edit `docker-compose.yml` and `docker-compose.dev.yml`. Set the new PostgreSQL version:

```yaml
services:
  postgres:
    image: postgres:17
    # [...]
```

## Start the PostgreSQL database server
```shell
$ docker compose up -d postgres

[+] Running 3/3
 ✔ Network sparklemuffin_default         Created
 ✔ Volume "sparklemuffin_postgres-data"  Created
 ✔ Container sparklemuffin-postgres-1    Started
```

## Restore the PostgreSQL database
```shell
$ make pgrestore

# docker compose exec -T postgres pg_restore -U sparklemuffin --dbname sparklemuffin < dump/sparklemuffin.sql.zst
```

```shell
$ make pgreindex

# docker compose exec postgres psql -U sparklemuffin -d sparklemuffin -c "REINDEX DATABASE sparklemuffin;"
REINDEX

# docker compose exec postgres psql -U sparklemuffin -d sparklemuffin -c "ALTER DATABASE sparklemuffin REFRESH COLLATION VERSION;"
ALTER DATABASE
```

## Verification
```shell
$ make psql

# docker compose exec postgres psql -U sparklemuffin

psql (17.5 (Debian 17.5-1.pgdg120+1))
Type "help" for help.

sparklemuffin=# SELECT COUNT(*) FROM bookmarks;

 count
-------
  5126
(1 row)
```

## Reference
### PostgreSQL documentation
- [pg_dump](https://www.postgresql.org/docs/current/app-pgdump.html) - Extract a PostgreSQL database into a script file or other archive file
- [pg_restore](https://www.postgresql.org/docs/current/app-pgrestore.html) - Restore a PostgreSQL database from an archive file created by `pg_dump`
- [psql](https://www.postgresql.org/docs/17/app-psql.html) - PostgreSQL interactive terminal
- [PostgreSQL 16 Release Notes](https://www.postgresql.org/docs/release/16.0/) - PostgreSQL 16 adds LZ4 and Zstandard compression to `pg_dump`

### Sparklemuffin database
- [Database](../reference/database.md)

### Articles
- [Is pg_dump a Backup Tool?](https://rhaas.blogspot.com/2024/10/is-pgdump-backup-tool.html), Robert Haas, 2024-10-15
