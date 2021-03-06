# goConsulKV
#### Console based go application for Consul key-Value store operations
----------

[![CircleCI](https://circleci.com/gh/iAviPro/goConsulKV/tree/master.svg?style=shield&circle-token=a0b171036fba85469fbfa175a73ed0e7223357ab)](https://app.circleci.com/pipelines/github/iAviPro)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/iAviPro/goConsulKV?tab=overview)](https://pkg.go.dev/github.com/iAviPro/goConsulKV@v1.0.1?tab=overview)

### About:
Consul by HashiCorp is service discovery & health-checking system. It also provides a Key-Value store for services, to be used as service or environment properties. This console based application builds on that key/value store.
If your development / testing environment uses single or multiple consuls & utilizes it's kv-store, then you can use this app to do operations on the key/value store.

### Get Started

Pre-requisite: Go 1.14 or above.

1. `git clone github.com/iAviPro/goConsulKV`
2. Update `./config/consulConfig.yml` or create your own yml config file.
3. `go build`
4. `./goConsulKV <$command> <$arguments>`

### Commands, their Arguments with details

| **__Command__** | **__Arguments__** | **__Details__**                                              |
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
|     sync        | -config            | Define consul config yml file. Default is `./config/consulConfig.yml` |
|                 | -replace           | ['true' / 'false'] Replaces the Value if Key(s) already exists. Default is false. |
|                 | -s                 | Define service name. Default is empty string.                |
|                 | -source (_Required_)    | Define source consul name as per config yml. Default is empty.    |
|                 | -target (_Required_)    | Define target consul name as per config yml. Default is empty.                |

`./goConsulKV <$command> -help` for further details

### Use Cases

1. Add / Update KV Pairs in one or multiple consul servers in single step.
2. Delete KV Pairs in one or multiple consul servers in single step.
3. Backup of one or multiple consul servers kv-store in single step.
4. Restore backup of one or multiple consul servers in single step.
5. Sync two consul KV store entirely or all KVs of a given service path.

### Sample Commands for Different Use-Case Scenarios

Update ./config/consulConfig.yml file with consul server details.  
   OR.  
Define your own config yml file and give absolute path to that file using `-config` argument in the command.  

**Add Key-Values to multiple consuls (dev / staging / prod environment):**  
>```./goConsulKV add -s serviceName1 -p "propKey1=propValue1 | propKey2=propValue2 | propKey3='propValue3'"```  

**Add Key-Values & Update Value if Key exists for multiple consuls (dev / staging / prod environment):**  
>```./goConsulKV add -s serviceName1 -replace true -p "propKey1=propValue1 | propKey2=propValue2 | propKey3='propValue3'"```  

**Add / Update Key-Values in single consul:**  
>```./goConsulKV add -n consulName1 -s serviceName1 -p "propKey1=propValue1 | propKey2=propValue2 | propKey3='propValue3'"```  

**Delete Key-Values to multiple consuls (dev / staging / prod environment):**  
>```./goConsulKV delete -s serviceName1 -replace true -p "propKey1=propValue1 | propKey2=propValue2 | propKey3='propValue3'"```  

**Delete Key-Values in single consul:**  
>```./goConsulKV delete -n consulName1 -s serviceName1 -p "propKey1=propValue1 | propKey2=propValue2 | propKey3='propValue3'"```  

**Backup Key-Value Store of multiple Consuls:**  
>```./goConsulKV backup -save "/home/users/username/consul_backup"```  

**Backup Key-Value Store of single Consul:**  
>```./goConsulKV backup -n consulName1 -save "/home/users/username/consul_backup"```  

**Restore Key-Value Store of a Consul for all KVs:**  
>```./goConsulKV restore -n consulName1 -file "/home/users/username/consul_backup/consulName1.json"```  

**Restore Key-Value Store of a single service in a Consul:**  
>```./goConsulKV restore -n consulName1 -file "/home/users/username/consul_backup/consulName1.json" -s serviceName1```  

**Sync all KVs of Source Consul KV-store with Target Consul KV-store without replacing existing Key-Values in target:**
>```./goConsulKV sync -source consulName1 -target consulName2```

**Sync all KVs of Source Consul KV-store with Target Consul KV-store and replacing existing Key-Values in target:**
>```./goConsulKV sync -source consulName1 -target consulName2 -replace true```

**Sync KVs of a given service from Source Consul KV-store to Target Consul KV-store for same service without replace:**
>```./goConsulKV sync -source consulName1 -target consulName2 -s serviceName1```

**Sync KVs of a given service from Source Consul KV-store to Target Consul KV-store for same service with replace:**
>```./goConsulKV sync -source consulName1 -target consulName2 -s serviceName1 -replace true```

*Note: The argument values are place holders in the above sample command*  

> Pro Tip: You can use the commands in Jenkins Job as well to create cron backups, restore, update or delete consul KVs.

------
Primary dependency: https://godoc.org/github.com/hashicorp/consul/api
