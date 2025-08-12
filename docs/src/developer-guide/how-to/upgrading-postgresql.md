# Upgrading the major version of the local PostgreSQL server
When running the PostgreSQL database server as a Docker container, the most straightforward approach
to upgrade the major version of PostgreSQL is to:

1. Start the server
2. Dump the database to a local file using `pg_dump`
3. Stop the server and destroy the Docker volume containing the PostgreSQL server data
4. Bump the PostgreSQL version
5. Start the server
6. Restore the database from the local dump file using `pg_restore`


## Start the PostgreSQL database server
```shell
$ docker compose up -d postgresql

[+] Running 3/3
 ✔ Network sparklemuffin_default         Created
 ✔ Volume "sparklemuffin_postgres-data"  Created
 ✔ Container sparklemuffin-postgres-1    Started
```

## Dump the database
```shell
$ make pgdump

# mkdir -p dump
# docker compose exec postgres pg_dump -U sparklemuffin sparklemuffin --format tar > dump/sparklemuffin.sql.tar
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
Edit `docker-compose.yml` and `docker-compose.dev.yml` to set the desired version of the PostgreSQL server:

```yaml
services:
  postgres:
    image: postgres:17
    # [...]
```

## Start the PostgreSQL database server
```shell
$ docker compose up -d postgresql

[+] Running 3/3
 ✔ Network sparklemuffin_default         Created
 ✔ Volume "sparklemuffin_postgres-data"  Created
 ✔ Container sparklemuffin-postgres-1    Started
```

## Restore the PostgreSQL database
```shell
$ make pgrestore

# docker compose exec -T postgres pg_restore -U sparklemuffin --dbname sparklemuffin < dump/sparklemuffin.sql.tar
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

### Sparklemuffin database
- [Database](../reference/database.md)

### Articles
- [Is pg_dump a Backup Tool?](https://rhaas.blogspot.com/2024/10/is-pgdump-backup-tool.html), Robert Haas, 2024-10-15
