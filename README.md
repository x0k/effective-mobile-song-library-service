# Song Library Service

## Usage

Define the following environment variables:

- `MUSIC_INFO_SERVICE_ADDRESS`
- `PG_CONNECTION_URI`

Run the application: `go run cmd/app/main.go`

## Documentation

[Swagger](/api/swagger.json)

[Useful commands](/mkfile)

[Development environment](/flake.nix)

## Feedback

### From reviewer

- A band can have many songs and storing the data in a heap in one table is a violation of database normalization

### From myself

- It was fun to work on the generalized expression lexer and the generalized filter for PostgreSQL tables implemented based on it!

- It looks like I missed checking for records in the update handler.
I should have checked `RowsAffected` and returned a 404 if there were no records.
