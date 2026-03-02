# slinky
Opinionated short link service written in Go 🔗

While existing shorteners like <https://bit.ly> are ubiquitous, they tend to lack some "creature comforts" that would improve the experience for people using and sharing short links:
* Links are case-sensitive and full of confusing characters, making it difficult to add to event posters or share verbally
* Many of the existing link shorteners are generic, and there's no way to validate where links come from
* Many online communities run on Discord, and the embed experience for redirected Discord invite links leaves much to be desired

slinky looks to solve these issues for smaller communities that are looking to share links with their members

## Goals
* Creation and use of random and vanity short links
* Optional authentication for link creation
* Case-insensitive links for easy display
* Limited alphabet for random links to prevent confusion (`0` vs `O`, etc.)
* Standards-compliant QR code generation for use on event posters
* Custom embeds for Discord invite links
* Run on a single VPS or a Kubernetes cluster

## Non-Goals
* Infinite short links: this URL shortener is designed for use by small communities like universities or social clubs
* Account management: accounts should be managed by an upstream service, which creates a JWT that can authenticate with slinky on the user's behalf

## Usage
### Database
slinky supports storing link information in PostgreSQL databases

We recommend creating a new database, then executing [slinky.sql](slinky.sql) to populate it with the necessary tables
```bash
createdb slinky
psql slinky < slinky.sql
```

### Endpoints
slinky supports the following endpoints:
* `GET /{link}` – Redirects a user to a registered short link
  * When Discord's crawler attempts to navigate to these URLs, and the URL leads to a valid Discord invite, it will be presented with a custom embed that shows information about the destination server
* `GET /{link}.png` – Generates a QR code for the short link
  * Accepts a parameter, `?size`, which is the width in pixels of the desired QR code, up to `1024`
* `POST /api/links/` – Generates a short link
  * If `SLINKY_REQUIRE_AUTH` is enabled, a Bearer `Authorization` header with a valid JWT is required to use this endpoint. *See [Authorization](#authorization) for more information*
  * Accepts a JSON body:
    ```json
    {
        "url": "<the long URL to shorten>",
        "vanity_url": "[optional] <a vanity path to use instead of a randomly generated one>"
    }
    ```

### Authorization
slinky supports optional [JWT](https://jwt.io) authorization for creating short links

If authorization is enabled, slinky will expect a Bearer `Authorization` header with a valid [HMAC-SHA256](https://en.wikipedia.org/wiki/HMAC)-signed JWT

This JWT must include the `sub` (subject) claim as well as a `roles` claim, which is a list that can include any of the following:
* `"Creator"` – Allows creation of "generated" short links, which use a random series of characters, like `/zkmzdd`, which the user cannot control
* `"VanityCreator"` – Allows creation of "vanity" short links, like `/info`

### Configuration
slinky is a [Twelve Factor](https://www.12factor.net/) application that is configured with the following environment variables
| Variable | Description | Default |
|----------|-------------|---------|
| `SLINKY_BASE_URL` | The URL to use as the base for QR code generation. This should match whichever base URL you expect your users to use to navigate to your shortener | |
| `SLINKY_LISTEN_ADDR` | The host and port where the slinky web server should listen | `:8080` |
| `SLINKY_DATABASE_URI` | The [connection URI](https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING-URIS) for a PostgreSQL database to store short link information | `postgresql://localhost:5432/slinky?sslmode=disable` |
| `SLINKY_LINK_SIZE` | The character length of randomly generated short links | `6` |
| `SLINKY_MAX_RETRIES` | The number of times to attempt to generate a random short link in the event of a name conflict | `3` |
| `SLINKY_DISCORD_POWERED_BY` | The text to put in the "Powered By" section of Discord invite embeds | `Powered by slinky 🔗` |
| `SLINKY_DISCORD_EMBED_COLOR` | The color of Discord invite embeds | `#ef6a9a` |
| `SLINKY_REQUIRE_AUTH` | Whether or not this instance should require authorization to create links (see [Authorization](#authorization)) | `true` |
| `SLINKY_JWT_HMAC_SECRET` | HMAC-SHA256 secret used for verifying signed JWTs | |

### Running
slinky is a self-contained executable. After you have set the appropriate environment variables, you can run it with;
```bash
./slinky
```