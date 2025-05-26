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

## Unit Tests

The project includes a suite of unit tests to ensure code quality and correctness. Due to complexities encountered in the build and test environment during development, not all tests are currently passing, and precise code coverage metrics could not be obtained. However, the tests were designed with high coverage in mind.

The tests are organized as follows:

*   **`config/consulDetails_test.go`**: Contains tests for parsing the Consul configuration YAML file (`consulConfig.yml`). It verifies that the YAML is correctly unmarshalled into the configuration structs and that the mapping of configurations to a usable map structure (`GetConsulConfigMap`) is accurate. Tests include scenarios for valid, invalid, and non-existent configuration files.

*   **`src/helpers_test.go`**: Provides comprehensive tests for various utility functions used throughout the application. This includes:
    *   Parsing and cleaning key-value pair strings.
    *   Conversion of data structures to and from JSON (for backup/restore).
    *   File I/O operations for creating backup files and reading them.
    *   Validation of file paths.
    *   Logic for comparing and filtering lists of key-value pairs for synchronization.
    *   String manipulation utilities.

*   **`src/consulwrapper_test.go`**: Focuses on testing the core logic for interacting with the Consul Key-Value store. These tests were designed to use mock implementations of the Consul API client. The intended coverage includes:
    *   Connecting to Consul.
    *   Performing basic CRUD (Create, Read, Update, Delete) operations on KV pairs (`putKV`, `getKV`, `deleteKV`, `listAllKV`).
    *   Higher-level operations like `AddKVToConsul`, `DeleteKVFromConsul`, `BackupConsulKV`, `RestoreConsulKV`, and `SyncConsulKVStore`, verifying their logic with different input parameters and simulated Consul responses.

*   **`src/executor_test.go`**: Tests the command-line interface logic. This involves:
    *   Simulating various command-line arguments for each subcommand (`add`, `delete`, `backup`, `restore`, `sync`).
    *   Verifying that the correct functions from `consulwrapper.go` are called with the appropriate parameters.
    *   Testing argument validation, including missing arguments and help flag invocation. Mocking was intended for `os.Exit` calls to verify error exits.

*   **`main_test.go`**: A simple test to ensure that the `main()` function in `main.go` correctly calls the main execution function (`src.ExecuteGoConsulKVFunc`).

### Running the Tests

To run the unit tests, navigate to the root directory of the project and execute the following command:

```bash
go test ./...
```

This command will discover and run all test files in the current directory and its subdirectories.

*(Note: Due to the aforementioned environment complexities, some tests may not pass or execute correctly at this time.)*

### Test Coverage

The goal for these unit tests was to achieve near 100% code coverage. However, due to the technical difficulties encountered in the build and test environment, which prevented the tests from running to completion reliably, accurate code coverage metrics could not be generated at the time of this writing.

Once the environment issues are resolved and the tests are passing, code coverage can be measured using Go's built-in tools:

1.  Generate a coverage profile:
    ```bash
    go test -coverprofile=coverage.out ./...
    ```
2.  View the coverage report in HTML format:
    ```bash
    go tool cover -html=coverage.out
    ```
    Or, to get a summary by function in the terminal:
    ```bash
    go tool cover -func=coverage.out
    ```

------
Primary dependency: https://godoc.org/github.com/hashicorp/consul/api
