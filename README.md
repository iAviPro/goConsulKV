# goConsulKV
Console based solution for handling Consul key-Value store

### About:
Consul is service discovery & health-checking system. It also provides a Key-Value store for services, to be used as service or environment properties. This console based application builds on that key/value store.
If your development / testing environment uses single or multiple consuls & utilizes it's kv-store, then you can use this app to do operations  on the key/value store.

### Use Cases

1. Add / Update KV Pairs in one or multiple consul servers in single step.
2. Delete KV Pairs in one or multiple consul servers in single step.
3. Take backups of one or multiple consul servers kv-store in single step.
4. Restore backup of one or multiple consul servers in single step.

### Get Started

Pre-requisite: Go 1.14 or above.

1. `git clone github.com/iAviPro/goConsulKV`
2. `go build`
3. Update `./config/consulConfig.yml` or create your own yml config file.
4. `./goConsulKV <$command> <$arguments>`

### Commands and their params

| **__Command__** | **__Parameters__** | **__Details__**                                              |
| :-------------: | ------------------ | :----------------------------------------------------------- |
|       add       | -config            | Define consul config yml file. Default is `./config/consulConfig.yml` |
|                 | -n                 | Define consul name as per config yml. Default is empty string, which updates all consuls in config yml |
|                 | -p (_Required_)    | Define '\|' separated properties. Default is empty string    |
|                 | -replace           | ['true' / 'false'] Replaces the Value if Key already exists. Default is false. (default "false") |
|                 | -s                 | Define service name. Default is empty string.                |
|                 | -t                 | Define valid token. Default is empty string                  |
|     delete      | -config            | Define consul config yml file. Default is `./config/consulConfig.yml` |
|                 | -n                 | Define consul name as per config yml. Default is empty string, which updates all consuls in config yml |
|                 | -p (_Required_)    | Define '\|' separated properties. Default is empty string    |
|                 | -s (_Required_)    | Define service name. Default is empty string.                |
|                 | -t                 | Define valid token. Default is empty string                  |
|     backup      | -config            | Define consul config yml file. Default is `./config/consulConfig.yml` |
|                 | -n                 | Define consul name as per config yml. Default is empty string, which updates all consuls in config yml |
|                 | -save              | Define absolute directory path (without trailing '/') to save the backup file, given consul name will be the json  file name. Default is empty string, which backs-up at `./backup/${consul-name}.json` |
|                 | -t                 | Define valid token. Default is empty string                  |
|     restore     | -config            | Define consul config yml file. Default is `./config/consulConfig.yml` |
|                 | -file              | Define absolute file path to recovery json file. Default is empty string, which tries to recover from `./backup/${consul-name}.json` |
|                 | -n (_Required_)    | Define consul name as per config yml. Default is empty string. |
|                 | -s                 | Define service name. Default is empty string.                |
|                 | -t                 | Define valid token. Default is empty string                  |

`./goConsulKV <$command> -help` for further details
