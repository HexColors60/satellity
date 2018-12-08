## Introduce

Godiscourse is a 100% open source, free, go back-end discourse like forum, build from `hello world`. You can review it [https://suntin.com](https://suntin.com)

## Environment

1. GO version go1.10.2 darwin/amd64
2. postgres (PostgreSQL) 10.5

## Usage

1. Copy `config/test.cfg` to `config/config.go`.
2. Prepare database, you can find database schema in `models/schema.sql`, find the database config in `config/config.cfg`.
3. Run back-end `go build` and `./godiscourse`.
4. All front-end is under `web` directory, install dependence `npm install`, start service `npm run dev`.

## License

Released under the MIT license
