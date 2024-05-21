# URL shortener written in go
I created this url shortener as a small example project to learn go.

This project uses turso as the database to save the links.

## Usage

Compile the project and start it with the following environment variables:

```bash
export TURSO_AUTH_TOKEN="YOUR_AUTH_TOKEN"
export TURSO_DATABASE_URL="YOUR_DATABASE_URL"
```

You can create a turso DB using the turso [guide](https://docs.turso.tech/quickstart).

## Development

For development start the project using
```bash
go run main.go
```

For building:
```bash
go build main.go
```
