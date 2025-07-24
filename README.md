# `esq` - Elasticsearch Query CLI 🕵️‍♂️

**`esq`** is a powerful and simple command-line interface (CLI) for querying your Elasticsearch cluster. It's designed to make searching data easy, whether you prefer KQL, Lucene, or the full Elasticsearch Query DSL. You can also format and process the output on the fly.

## ✨ Features

  * **Flexible Querying**: Use the query language you're most comfortable with.
      * Kibana Query Language (**KQL**) via `--kql`
      * Lucene query syntax via `--lucene`
      * Full Elasticsearch Query **DSL** via `--dsl` or from a file with `--query-file`
  * **Time-Range Filtering**: Easily narrow your search to a specific time window using `--from` and `--to`.
  * **Powerful Output Processing**:
      * Format results as **JSON** or **text**.
      * Apply **`jq` expressions** directly to the output to reshape the JSON data.
      * Save results directly to a file.
  * **Flexible Configuration**: Configure `esq` via command-line flags, environment variables (e.g., `ESQ_NODE`), or a YAML config file.
  * **Simple Authentication**: Connect to secure clusters using an **API Key** or **Username/Password**.

-----

## 🚀 Installation

You can install `esq` using `go`:

```sh
go install github.com/fa7ad/esq/cmd/esq@latest
```

## ⚙️ Configuration

`esq` can be configured in three ways, with the following order of precedence:

1.  **Command-line flags** (e.g., `--node ...`)
2.  **Environment variables** (e.g., `export ESQ_NODE=...`)
3.  **Configuration file**

By default, `esq` looks for a configuration file at `$HOME/.esq.yaml`.

### Example `.esq.yaml`

```yaml
node: "http://localhost:9200"
index: "my-logs-*"
output: "json"
# api-key: "your_base64_api_key"
# username: "elastic"
# password: "changeme"
```

-----

## 💡 Usage

The only required flags are `--node` and `--index`. You must also provide one query flag: `--kql`, `--lucene`, `--dsl`, or `--query-file`.

### Examples

**1. Basic KQL Query**
Search for successful events for a specific user.

```sh
esq --node http://localhost:9200 --index 'my-logs-*' --kql "status:success and user:john"
```

**2. DSL Query from a File**
Execute a complex query stored in a JSON file and format the output as pretty-printed JSON.

```sh
esq -n https://es.example.com -i orders --query-file my_query.json -o json
```

**3. Time Range and `jq` Processing**
Find errors from the last hour and use `jq` to extract just the document ID and source.

```sh
esq -n http://localhost:9200 -i my-index \
  --kql "log.level:error" \
  --from "now-1h" --to "now" \
  -o json --jq ".hits | map({id: ._id, source: ._source})"
```

**4. Authentication**
Authenticate using an API key.

```sh
esq -n https://es.example.com -i metrics \
  --kql "cpu.usage > 90" \
  --api-key "your_base64_api_key"
```

Or use username and password, which can also be set via `ESQ_USERNAME` and `ESQ_PASSWORD` environment variables.

```sh
esq -n https://secure-es:9200 -i audit-logs \
  --username elastic --password changeme \
  --kql "event.action:login_failed"
```

### All Flags

```
A CLI tool to query Elasticsearch.

esq can be configured using command-line flags, environment variables (prefixed with ESQ_),
or a configuration file (e.g., $HOME/.esq.yaml).

Usage:
  esq [flags]

Flags:
      --api-key string       Elasticsearch API Key for authentication.
      --config string        config file (default is $HOME/.esq.yaml)
      --dsl string           Elasticsearch Query DSL JSON string.
  -f, --query-file string    Path to a file containing the Elasticsearch Query DSL (JSON).
      --from string          Start time (ISO8601 or ES-relative like 'now-1d').
  -h, --help                 help for esq
  -i, --index string         Elasticsearch index pattern.
  -j, --jq string            Apply a jq expression to the output.
      --kql string           Kibana Query Language (KQL) query string.
      --lucene string        Lucene query string.
  -n, --node string          Elasticsearch node URL.
  -o, --output string        Output format (choices: json, text) (default "text")
      --output-file string   Write output to a file instead of stdout.
      --password string      Password for basic authentication.
  -s, --size int             Number of results to return. (default 100)
      --to string            End time (ISO8601 or ES-relative like 'now').
      --username string      Username for basic authentication.
  -v, --version              version for esq
```