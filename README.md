# goConsulKV
Console based solution for handling Consul key-Value store

### About:
Consul is service discovery & health-checking system. It also provides a Key-Value store for services, to be used as service or environment properties. This console based application builds on that key/value store.
If your development / testing environment uses single or multiple consuls & utilizes it's kv-store, then you can use this app to do operations  on the key/value store.

### Use Cases:

1. Add / Update KV Pairs in one or multiple consul servers in single step.
2. Delete KV Pairs in one or multiple consul servers in single step.
3. Take backups of one or multiple consul servers kv-store in single step.
4. Restore backup of one or multiple consul servers in single step. [Work In progress]

### Get Started

Pre-requisite: Go 1.14 or above.

1. `git clone github.com/iAviPro/goConsulKV`
2. `go build`
3. Update `./config/consulConfig.yml` or create your own yml config file.
4. `./goConsulKV <$command> <$arguments>`
