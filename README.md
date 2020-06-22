# goConsulKV
Console based go application for Consul key-Value store operations

### About:
Consul is service discovery & health-checking system. It also provides a Key-Value store for services, to be used as service or environment properties. This console based application builds on that key/value store.
If your development / testing environment uses single or multiple consuls & utilizes it's kv-store, then you can use this app to do operations on the key/value store.

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
|                 | -s (_Required_)    | Define service name. Default is empty string.                |
|     delete      | -config            | Define consul config yml file. Default is `./config/consulConfig.yml` |
|                 | -n                 | Define consul name as per config yml. Default is empty string, which updates all consuls in config yml |
|                 | -p (_Required_)    | Define '\|' separated properties. Default is empty string    |
|                 | -s (_Required_)    | Define service name. Default is empty string.                |
|     backup      | -config            | Define consul config yml file. Default is `./config/consulConfig.yml` |
|                 | -n                 | Define consul name as per config yml. Default is empty string, which updates all consuls in config yml |
|                 | -save              | Define absolute directory path (without trailing '/') to save the backup file, given consul name will be the json  file name. Default is empty string, which backs-up at `./backup/${consul-name}.json` |
|     restore     | -config            | Define consul config yml file. Default is `./config/consulConfig.yml` |
|                 | -file              | Define absolute file path to recovery json file. Default is empty string, which tries to recover from `./backup/${consul-name}.json` |
|                 | -n (_Required_)    | Define consul name as per config yml. Default is empty string. |
|                 | -s                 | Define service name. Default is empty string.                |

`./goConsulKV <$command> -help` for further details

### Use Cases

1. Add / Update KV Pairs in one or multiple consul servers in single step.
2. Delete KV Pairs in one or multiple consul servers in single step.
3. Take backups of one or multiple consul servers kv-store in single step.
4. Restore backup of one or multiple consul servers in single step.

#### Sample Commands for Different Use-Case Scenarios

Update ./config/consulConfig.yml file with consul server details.
   OR
Define your own config yml file and give absolute path to that file using `-config` parameter in the command.

**Add Key-Values to multiple consuls (dev / staging / prod environment):**
```./goConsulKV add -s serviceName1 -p "propKey1=propValue1 | propKey2=propValue2 | propKey3='propValue3'"```

**Add Key-Values & Update Value if Key exists for multiple consuls (dev / staging / prod environment):**
```./goConsulKV add -s serviceName1 -replace true -p "propKey1=propValue1 | propKey2=propValue2 | propKey3='propValue3'"```

**Add / Update Key-Values in single consuls:**
```./goConsulKV add -n consulName1 -s serviceName1 -p "propKey1=propValue1 | propKey2=propValue2 | propKey3='propValue3'"```

**Delete Key-Values to multiple consuls (dev / staging / prod environment):**
```./goConsulKV delete -s serviceName1 -replace true -p "propKey1=propValue1 | propKey2=propValue2 | propKey3='propValue3'"```

**Delete Key-Values in single consuls:**
```./goConsulKV delete -n consulName1 -s serviceName1 -p "propKey1=propValue1 | propKey2=propValue2 | propKey3='propValue3'"```

**Backup Key-Value Store of multiple Consuls:**
```./goConsulKV backup -save "/home/users/username/consul_backup"```

**Backup Key-Value Store of single Consuls:**
```./goConsulKV backup -n consulName1 -save "/home/users/username/consul_backup"```

**Restore Key-Value Store of a Consuls for all KVs:**
```./goConsulKV backup -n consulName1 -file "/home/users/username/consul_backup/consulName1.json"```

**Restore Key-Value Store of a single service in a Consuls :**
```./goConsulKV backup -n consulName1 -file "/home/users/username/consul_backup/consulName1.json" -s serviceName1```

> Tip: You can use the commands in Jenkins Job as well to create cron backups, restore, update or delete consul KVs.