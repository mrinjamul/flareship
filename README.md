<div align="center">
  <h1><code>flareship</code></h1>
  <p>
    <strong>📦 A CLI to sync domains from local to Cloudflare.</strong>
  </p>
</div>

## Installation

### From source

If you want to build `flareship` from source, you need Go 1.20 or
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
FLARESHIP_DOMAINS="example.com,myapp.io"
FLARESHIP_CF_TOKENS="token1,token2"
FLARESHIP_ZONE_IDS="zone1,zone2"
FLARESHIP_RECORD_FILES="example_com.json,myapp_io.json"
FLARESHIP_RESTRICTED_FILES="restricted_example_com.json,restricted_myapp_io.json"
FLARESHIP_ALLOWED_TYPES="A,CNAME;A,CNAME"
```

Available envs:

- `FLARESHIP_ZONE_IDS`: Cloudflare zone id
- `FLARESHIP_CF_TOKENS`: Cloudflare API key
- `FLARESHIP_DOMAINS`: Top level domain names
- `FLARESHIP_RECORD_FILES`: Path to file with domains
- `FLARESHIP_RESTRICTED_FILES`: Path to file with restricted domains
- `FLARESHIP_ALLOWED_TYPES`: Allowed DNS Types for these domains

or

- `CONFIG_FILE`: location to configuration file (optional)

## Configurations

Use environment variables to configure the CLI.

or you can use configuaration file `flareship.json`

Create config file interactively:

```json
flareship init
```

## Usage

`flareship` is a CLI to sync domains from local to Cloudflare.

```
flareship CLI

Usage:
  flareship [flags]
  flareship [command]

Available Commands:
  backup      backup DNS records to file.
  completion  Generate the autocompletion script for the specified shell
  fmt         format the records
  help        Help about any command
  init        Initialize config and empty records
  list        list all records from remote/local
  sync        sync with remote DNS.
  version     prints version.

Flags:
  -c, --config string   specify config file location
  -h, --help            help for flareship

Use "flareship [command] --help" for more information about a command.
```

`flareship fmt --check` will check if the records are ok.

```
    format the records

    Usage:
    flareship fmt [flags]

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
  flareship sync [flags]

Flags:
      --domain string   specify the domain name
      --dry-run         dry run the sync
  -h, --help            help for sync
```

`flareship list` will list all records from remote/local.

```
list all records from remote/local

Usage:
  flareship list [flags]

Flags:
  -h, --help          help for list
  -l, --local         specify the target to list e.g. local
  -t, --type string   specify the types of records

```

`flareship backup` will export the records to a file.

```
backup DNS records to file.

Usage:
  flareship backup [flags]

Flags:
      --domain string   specify the domain name
  -h, --help            help for backup
  -t, --type string     specify the types of records

```

`flareship version` will print the version.

## License

- open sourced under [MIT license](LICENSE)
