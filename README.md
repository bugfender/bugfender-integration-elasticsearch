# Bugfender Integration with Elasticsearch

This sample project shows how to copy in real time all Bugfender logs to an Elasticsearch cluster.

## Usage

```
Usage of ./bugfender-integration-elasticsearch:
  -api-url="https://dashboard.bugfender.com": Bugfender API URL (only necessary for on-premises)
  -app-id=0: Bugfender app ID (mandatory)
  -client-id="": OAuth client ID to connect to Bugfender (mandatory)
  -client-secret="": OAuth client secret to connect to Bugfender (mandatory)
  -config="": path to config file
  -console-output=false: Print logs to console instead of Elasticsearch (for debugging)
  -es-index="": Elasticsearch index to write to (default: logs)
  -es-nodes="": List of Elasticsearch nodes (multiple nodes can be specified, separated by spaces)
  -es-password="": Password to connect to Elasticsearch
  -es-username="": Username to connect to Elasticsearch
  -insecure-skip-tls-verify=false: Skip TLS certificate verification (insecure)
  -retries=10: Number of times to retry on errors before exiting. 0 = never give up.
  -state-file="": File to restore and save state, to resume sync (recommended)
  -verbose=false: Verbose messages
  ```

A typical example on how to run this tool would be:

    ./bugfender-integration-elasticsearch -app-id=1234 -client-id=your_client_id -client-secret=your_client_secret -state-file state.json -es-index logs -es-nodes http://127.0.0.1:9200

An example Elasticsearch instance can be run with the provided `docker-compose.yml` file.

If you would like to test this tool without an Elasticsearch, you can dump the logs to console:

    ./bugfender-integration-elasticsearch -app-id=1234 -client-id=your_client_id -client-secret=your_client_secret -state-file state.json -console-output

## Writing integrations with other databases

Integrations with different databases can be written if desired, by implementing the `integration.LogWriter` interface.

An example of such integration is the `pkg/dummy` package, which dumps the received logs to the console.

## Contributing

Contributions to this project are welcome, please feel free to open a pull request.

For bugs, you can open an issue in Github or submit a pull request.

For security vulnerabilities, [contact our security staff](https://support.bugfender.com/en/articles/580923-security-contact).