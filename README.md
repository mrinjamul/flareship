<div align="center">
  <h1><code>flareship</code></h1>
  <p>
    <strong>📦 A CLI to sync domains from local to Cloudflare.</strong>
  </p>
</div>

## Installation

### From source

If you want to build `flareship` from source, you need Go 1.16 or
higher. You can then use `go build` to build everything:

```
git clone https://github.com/mrinjamul/flareship.git
cd flareship
go mod download
make install
```

#### From prebuilt binaries:

You can download prebuilt binaries from the [Release page](https://github.com/mrinjamul/flareship/releases),

#### Via Homebrew

```
    brew tap mrinjamul/main
    brew install flareship
```

## Prerequisites

we need to have a Cloudflare account and API key.

and we need to have a top level domain.

set up `.env` file with content,

```
CF_ZID="your-zone-id"
CF_TOK="your-api-key"
```

Available envs:

- `CF_ZID`: Cloudflare zone id (required)
- `CF_TOK`: Cloudflare API key (required)
- `DOMAIN_NAME`: Top level domain name (optional)
- `RECORD_FILE`: Path to file with domains (optional)
- `RESTRICTED_FILE`: Path to file with restricted domains (optional)

or

- `CONFIG_FILE`: location to configuration file (optional)

## Configurations

Use environment variables to configure the CLI.

or you can use configuaration file `$HOME/.mrinjamulcli.json`

Sample config file:

```json
{
  "cf_token": "your-api-key",
  "zone_id": "your-zone-id",
  "domain_name": "your-domain.com",
  "record_file": "records.json",
  "restricted_file": "restricted.json"
  "record_type": ["A", "CNAME"]
}
```

## Usage

`flareship` is a CLI to sync domains from local to Cloudflare.

```
    mrinjamul.in CLI

    Usage:
    mrinjamul [flags]
    mrinjamul [command]

    Available Commands:
    completion  Generate the autocompletion script for the specified shell
    export      export DNS records to file.
    fmt         format the records
    help        Help about any command
    sync        sync with remote DNS.
    version     prints version.

    Flags:
    -h, --help   help for mrinjamul

    Use "mrinjamul [command] --help" for more information about a command.

```

`flareship fmt --check` will check if the records are ok.

```
    format the records

    Usage:
    mrinjamul fmt [flags]

    Flags:
    -c, --check           checks if the records has for errors
        --domain string   specify the domain name
    -f, --file string     specify the records file
    -h, --help            help for fmt

```

`flareship sync` will sync the records from local to remote.

```
    sync with remote DNS.

    Usage:
    mrinjamul sync [flags]

    Flags:
        --domain string   specify the domain name
        --dry-run         dry run the sync
    -f, --file string     specify the records file
    -h, --help            help for sync
    -p, --proxied         set all records proxied

```

`flareship export` will export the records to a file.

```
    export DNS records to file.

    Usage:
    mrinjamul export [flags]

    Flags:
        --domain string   specify the domain name
    -f, --file string     specify the export file
    -h, --help            help for export

```

`flareship version` will print the version.

```
    prints version.

    Usage:
    mrinjamul version [flags]

    Flags:
    -h, --help   help for version

```

## License

- open sourced under [MIT license](LICENSE)
