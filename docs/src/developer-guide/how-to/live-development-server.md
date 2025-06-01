# Live Development Server
## Prerequisites
- [GNU Make](https://www.gnu.org/software/make/)
- [Docker](https://docs.docker.com/) and [Docker Compose](https://docs.docker.com/compose/)
- [watchexec](https://github.com/watchexec/watchexec)
- [Go](https://go.dev/)
- [Node.js](https://nodejs.org/)

## Run a local development server
The helper Make targets will run a local development server, with:

- the PostgreSQL database running as a Docker container;
- the SparkleMuffin application running locally in development mode (via `go run`).

The application server will be reloaded every time a source file is changed on the disk,
thanks to `watchexec`.

For more information about how services are configured and started, see:

- the [Makefile](https://github.com/virtualtam/sparklemuffin/blob/main/Makefile);
- the [docker-compose.dev.yml](https://github.com/virtualtam/sparklemuffin/blob/main/docker-compose.dev.yml) Docker Compose configuration.

Run a local development server:

```shell
$ make live

== Downloading frontend assets
cd internal/http/www/assets && npm ci

added 13 packages, and audited 14 packages in 623ms

found 0 vulnerabilities
== Building frontend assets
cd internal/http/www/assets && go run main.go

  ../static/www.min.css                          762.4kb
  ../static/fa-solid-900-PJNKLK6W.ttf            416.1kb
  ../static/FiraCode-VF-AEJJ5BCX.ttf             279.6kb
  ../static/Exo2-VariableFont_wght-Q2QZZPLQ.ttf  276.9kb
  ../static/fa-brands-400-R2XQZCET.ttf           205.9kb
  ../static/fa-solid-900-5ZUYHGA7.woff2          154.5kb
  ../static/fa-brands-400-Q3XCMWHQ.woff2         115.9kb
  ../static/fa-regular-400-XUOPSR7E.ttf           66.5kb
  ../static/fa-regular-400-QSNYFYRT.woff2         24.9kb
  ../static/fa-v4compatibility-YY67RJWG.ttf       10.6kb
  ../static/fa-v4compatibility-LFEHZI6Y.woff2      4.7kb

⚡ Done in 38ms
2025/06/01 21:06:18 copied node_modules/awesomplete/awesomplete.min.js to ../static/awesomplete.min.js
2025/06/01 21:06:18 copied node_modules/easymde/dist/easymde.min.js to ../static/easymde.min.js
2025/06/01 21:06:18 copied favicons/android-chrome-192x192.png to ../static/android-chrome-192x192.png
2025/06/01 21:06:18 copied favicons/android-chrome-512x512.png to ../static/android-chrome-512x512.png
2025/06/01 21:06:18 copied favicons/apple-touch-icon.png to ../static/apple-touch-icon.png
2025/06/01 21:06:18 copied favicons/favicon-16x16.png to ../static/favicon-16x16.png
2025/06/01 21:06:18 copied favicons/favicon-32x32.png to ../static/favicon-32x32.png
2025/06/01 21:06:18 copied favicons/favicon.ico to ../static/favicon.ico
2025/06/01 21:06:18 copied favicons/site.webmanifest to ../static/site.webmanifest
== Starting database
docker compose -f docker-compose.dev.yml up --remove-orphans -d
[+] Running 2/2
 ✔ Network sparklemuffin_default       Created                                                                                                                                                                                                            0.0s
 ✔ Container sparklemuffin-postgres-1  Started                                                                                                                                                                                                            0.3s
== Watching for changes... (hit Ctrl+C when done)
[Running: go run ./cmd/sparklemuffin/ run]
2025-06-01T21:06:19+02:00 INF configuration: no file found config_paths=["/etc","/home/dev/.config","."]
2025-06-01T21:06:19+02:00 INF database: successfully created connection pool database_addr=localhost:15432 database_driver=pgx database_name=sparklemuffin
2025-06-01T21:06:19+02:00 INF global: setting up services log_level=info version=devel
2025-06-01T21:06:19+02:00 INF feeds: synchronization scheduler started interval_seconds=3600000
2025-06-01T21:06:19+02:00 INF metrics: listening for HTTP requests metrics_addr=127.0.0.1:8081
2025-06-01T21:06:19+02:00 INF sparklemuffin: listening for HTTP requests http_addr=0.0.0.0:8080
```

Run a local development server, with the Go race detector enabled:

```shell
$ make live-race

== Starting database
docker compose -f docker-compose.dev.yml up --remove-orphans -d
[+] Building 0.0s (0/0)                                                                                                                                                                               docker:default
[+] Running 1/0
 ✔ Container sparklemuffin-postgres-1  Running                                                                                                                                                                  0.0s
== Watching for changes... (hit Ctrl+C when done)
2023-11-03T10:27:38+01:00 INF configuration: no file found config_paths=["/etc","/home/dev/.config","."]
2023-11-03T10:27:38+01:00 INF database: successfully created connection pool database_addr=localhost:15432 database_driver=pgx database_name=sparklemuffin
2023-11-03T10:27:38+01:00 INF global: setting up services log_level=info version=devel
2023-11-03T10:27:38+01:00 INF metrics: listening for HTTP requests metrics_addr=127.0.0.1:8081
2023-11-03T10:27:38+01:00 INF sparklemuffin: listening for HTTP requests http_addr=0.0.0.0:8080
```

## Run database migrations
```shell
$ make dev-migrate

go run ./cmd/sparklemuffin migrate
2023-11-03T10:31:53+01:00 INF configuration: no file found config_paths=["/etc","/home/dev/.config","."]
2023-11-03T10:31:53+01:00 INF database: successfully created connection pool database_addr=localhost:15432 database_driver=pgx database_name=sparklemuffin
2023-11-03T10:31:53+01:00 INF successfully opened database connection database_addr=localhost:15432 database_driver=pgx database_name=sparklemuffin
2023-11-03T10:31:53+01:00 INF migrate: the database schema is up to date database_addr=localhost:15432 database_driver=pgx
```

## Create a first administrator user
Create the user account with:

```shell
$ make dev-admin

go run ./cmd/sparklemuffin createadmin \
        --displayname Admin \
        --email admin@dev.local \
        --nickname admin
2023-11-03T10:34:50+01:00 INF configuration: no file found config_paths=["/etc","/home/dev/.config","."]
2023-11-03T10:34:50+01:00 INF database: successfully created connection pool database_addr=localhost:15432 database_driver=pgx database_name=sparklemuffin
2023-11-03T10:34:50+01:00 INF admin user successfully created email=admin@dev.local nickname=admin
Generated password: Qj3Qkeq4GpmEOrzjRv36VqVPQVymztbE4nlQ9u8KhjE=
```

Then open the application in your Web browser:

- access [http://localhost:8080](http://localhost:8080/);
- login using the generated credentials:
    - Email address: `admin@dev.local`
    - Password: use the password generated by the `make dev-admin` command


## Stop local services
```shell
$ docker compose stop

[+] Stopping 1/1
 ✔ Container sparklemuffin-postgres-1  Stopped
```

## Remove containers and application data
Stop and remove application containers (without removing data volumes):

```shell
$ docker compose down
```

Stop and remove application containers, and remove data volumes:

```shell
$ docker compose down -v
```
