package src

/*
	@author: Aviral Nigam
	@github: https://github.com/iAviPro
	@date: 15 Jun, 2020
*/

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/ConsulScale/config" // Assuming this is the correct import path
	"github.com/hashicorp/consul/api"
)

// ConsulKV is an interface abstracting Consul's KV operations for mocking.
type ConsulKV interface {
	Get(key string, q *api.QueryOptions) (*api.KVPair, *api.QueryMeta, error)
	List(prefix string, q *api.QueryOptions) (api.KVPairs, *api.QueryMeta, error)
	Put(p *api.KVPair, q *api.WriteOptions) (*api.WriteMeta, error)
	Delete(key string, q *api.WriteOptions) (*api.WriteMeta, error)
}

// ConsulManager manages Consul client connections and KV operations.
type ConsulManager struct {
	KVStore ConsulKV // Interface for KV operations
	Detail  *config.ConsulDetail
}

// ConnectConsul establishes a connection to a Consul server using details.
// Note: This function still returns *api.Client as it's about actual connection.
// NewConsulManager will then wrap this.
func ConnectConsul(detail *config.ConsulDetail) (*api.Client, error) {
	if detail == nil {
		return nil, errors.New("ConsulDetail is nil")
	}
	apiConfig := api.DefaultConfig()
	// Corrected field access based on actual config.ConsulDetail
	apiConfig.Address = detail.BaseURL // BaseURL likely contains the full address including scheme
	apiConfig.Datacenter = detail.DataCentre
	apiConfig.Token = detail.Token
	// Scheme might be part of BaseURL or implicitly http/https if BaseURL doesn't specify
	// If BaseURL does not include scheme, and apiConfig.Address needs only host:port,
	// then scheme needs to be parsed from BaseURL or set separately if available.
	// For now, assuming BaseURL is sufficient for apiConfig.Address.
	// If api.NewClient needs separate scheme, this might need further adjustment
	// based on how BaseURL is formatted. If BaseURL includes "http://", api.NewClient handles it.

	client, err := api.NewClient(apiConfig)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// ConnectConsulFunc points to the original ConnectConsul function.
var ConnectConsulFunc = ConnectConsul

// NewConsulManagerFunc points to the original NewConsulManager function.
var NewConsulManagerFunc = NewConsulManager

// NewConsulManager creates a new ConsulManager with a connected KV store.
func NewConsulManager(detail *config.ConsulDetail) (*ConsulManager, error) {
	if detail == nil {
		return nil, errors.New("ConsulDetail is nil for NewConsulManager")
	}
	// Use ConnectConsulFunc to allow mocking
	client, err := ConnectConsulFunc(detail)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Consul for manager: %w", err)
	}
	if client == nil { // Should be caught by ConnectConsulFunc, but good practice
		return nil, errors.New("failed to initialize Consul API client for manager")
	}
	return &ConsulManager{KVStore: client.KV(), Detail: detail}, nil
}

// PutKV stores a key-value pair in Consul.
// The key in p *api.KVPair is treated as a suffix to Detail.BasePath.
func (cm *ConsulManager) PutKV(p *api.KVPair) (bool, error) {
	if cm.KVStore == nil {
		return false, errors.New("Consul KVStore is nil")
	}
	fullKey := path.Join(cm.Detail.BasePath, p.Key)
	kvPairToPut := &api.KVPair{Key: fullKey, Value: p.Value, Flags: p.Flags}

	_, err := cm.KVStore.Put(kvPairToPut, nil)
	if err != nil {
		log.Printf("Error putting value to Consul for key %s: %s\n", fullKey, err)
		return false, err
	}
	fmt.Printf("Added KV: %s = %s \n", fullKey, p.Value) // Log with full key
	return true, nil
}

// GetKV retrieves a key-value pair from Consul.
// The key is treated as a suffix to Detail.BasePath.
func (cm *ConsulManager) GetKV(keySuffix string) (*api.KVPair, error) {
	if cm.KVStore == nil {
		return nil, errors.New("Consul KVStore is nil")
	}
	fullKey := path.Join(cm.Detail.BasePath, keySuffix)
	pair, _, err := cm.KVStore.Get(fullKey, nil)
	if err != nil {
		// Distinguish between "not found" and other errors if necessary,
		// but for now, any error is logged and returned.
		// A 404 from Consul typically results in `pair == nil` and `err == nil`.
		// The api.KV.Get method handles this.
		log.Printf("Error getting value from Consul for key %s: %s\n", fullKey, err)
		return nil, err
	}
	// If pair is nil and err is nil, it means key not found.
	return pair, nil
}

// DeleteKV deletes a key-value pair from Consul.
// The key is treated as a suffix to Detail.BasePath.
func (cm *ConsulManager) DeleteKV(keySuffix string) (bool, error) {
	if cm.KVStore == nil {
		return false, errors.New("Consul KVStore is nil")
	}
	fullKey := path.Join(cm.Detail.BasePath, keySuffix)
	_, err := cm.KVStore.Delete(fullKey, nil)
	if err != nil {
		log.Printf("Error deleting value from Consul for key %s: %s\n", fullKey, err)
		return false, err
	}
	fmt.Printf("Deleted Key: %s\n", fullKey)
	return true, nil
}

// ListAllKV lists all key-value pairs from Consul under the manager's BasePath.
func (cm *ConsulManager) ListAllKV() (api.KVPairs, error) { // Changed return type
	if cm.KVStore == nil {
		return nil, errors.New("Consul KVStore is nil")
	}
	// cm.Detail.BasePath is expected to be the prefix.
	// Ensure it ends with "/" if it's meant to be a folder prefix.
	// The List API call handles this appropriately.
	kvPairsResult, _, err := cm.KVStore.List(cm.Detail.BasePath, nil) // Renamed 'pairs' to 'kvPairsResult'
	if err != nil {
		fmt.Println("There was an error in ListAllKV() Operation")
		return nil, err
	}
	/* Debug: Print KV List */
	// for _, kv := range kvPairsResult {
	// 	fmt.Printf("%s = %s \n", kv.Key, kv.Value)
	// }
	return kvPairsResult, nil // Return api.KVPairs directly as per its signature
}

// AddKVToConsul : Allows only to add KV-pairs in multi-consul setup, if the KV is already existing it will not update KV
func AddKVToConsul(sn, cn, props, configPath, replace string) {
	configDetailsMap, err := CreateConsulDetailsFunc(configPath) // Use Func
	if err != nil {
		log.Fatalf("Failed to create consul details from config path %s: %v", configPath, err)
	}

	kvInputPairs := createKVPairs(props, sn) // These KVPairs have keys relative to serviceName (sn)

	processFunc := func(detail config.ConsulDetail) {
		manager, err := NewConsulManager(&detail)
		if err != nil {
			errConsulConnection(detail.ConsulName, err, false) // Use helper for connection error
			return
		}

		var addFail = make(map[string][]byte)
		for _, kv := range kvInputPairs { // kv.Key here is like "serviceName/actualKey" or just "actualKey" if sn is ""
			// The key for GetKV/PutKV should be relative to manager.Detail.BasePath
			// If sn (serviceName) is part of kv.Key, and BasePath is like "env/",
			// then PutKV will join them: "env/" + "serviceName/actualKey"
			// If createKVPairs already includes sn in the key, that's fine.
			// The methods on manager (PutKV, GetKV) expect keys relative to their BasePath.
			// So, if kv.Key is "myService/prop1" and BasePath is "dev/", PutKV makes it "dev/myService/prop1".
			// If sn is empty, createKVPairs makes keys like "prop1". PutKV makes "dev/prop1". This seems correct.

			if replace == "false" {
				existingPair, errGet := manager.GetKV(kv.Key) // kv.Key is the suffix
				if errGet != nil {
					fmt.Printf("<> -- Error checking key %s in Consul Server %s: %v -- <>\n", path.Join(manager.Detail.BasePath, kv.Key), detail.ConsulName, errGet)
					addFail[kv.Key] = kv.Value // Consider it failed if we can't check
					continue
				}
				if existingPair == nil { // Key does not exist
					_, errPut := manager.PutKV(&kv) // Pass the original kv directly
					if errPut != nil {
						fmt.Printf("<> -- Could not Add Key to Consul Server %s. KV: %s=%s. Error: %v -- <>\n", detail.ConsulName, kv.Key, string(kv.Value), errPut)
					}
				} else {
					addFail[kv.Key] = kv.Value
				}
			} else { // replace == "true"
				_, errPut := manager.PutKV(&kv)
				if errPut != nil {
					fmt.Printf("<> -- Could not Update Key in Consul Server %s. KV: %s=%s. Error: %v -- <>\n", detail.ConsulName, kv.Key, string(kv.Value), errPut)
				}
			}
		}

		if len(addFail) > 0 {
			fmt.Printf("\n <> -- Failed to Add/Replace Following KVs for Consul Server: %s (replace=%s) -- <> \n", detail.ConsulName, replace)
			for k, v := range addFail {
				fmt.Printf("Key: %s (Value: %s) - Either already exists (replace=false) or error during check/put.\n", k, string(v))
			}
			if replace == "false" {
				fmt.Println("\nNote:- Use --replace 'true' in the run command to replace values of existing keys. Use --help for more info.")
			}
		}
	}

	if cn == "" { // All consul configurations
		for _, detail := range configDetailsMap {
			processFunc(detail)
		}
	} else if IsConfigNameValidFunc(cn, configDetailsMap) { // Use Func
		processFunc(configDetailsMap[cn])
	} else {
		fmt.Println("<> -- Invalid Consul Server Name specified. Check config file and input. -- <>")
	}
}

/*
DeleteKVFromConsul : Deletes given keys from consul server
*/
func DeleteKVFromConsul(sn, cn, props, configPath string) {
	configDetailsMap, err := CreateConsulDetails(configPath)
	if err != nil {
		log.Fatalf("Failed to create consul details from config path %s: %v", configPath, err)
	}
	kvInputPairs := createKVPairs(props, sn) // Keys relative to serviceName

	processFunc := func(detail config.ConsulDetail) {
		manager, errMgr := NewConsulManager(&detail)
		if errMgr != nil {
			errConsulConnection(detail.ConsulName, errMgr, false)
			return
		}
		for _, kv := range kvInputPairs {
			// kv.Key is the key suffix for DeleteKV
			_, errDel := manager.DeleteKV(kv.Key)
			if errDel != nil {
				fmt.Printf("<> -- Could not Delete Key %s from Consul Server %s. Error: %v -- <>\n", path.Join(manager.Detail.BasePath, kv.Key), detail.ConsulName, errDel)
			}
		}
	}

	if cn == "" {
		for _, detail := range configDetailsMap {
			processFunc(detail)
		}
	} else if IsConfigNameValidFunc(cn, configDetailsMap) { // Use Func
		processFunc(configDetailsMap[cn])
	} else {
		fmt.Println("<> -- Invalid Consul Server Name specified. Check config file and input. -- <>")
	}
}

/*
BackupConsulKV : Function to take json backups of consul and save to a backup file
*/
func BackupConsulKV(cn, configPath, backupFilePath string) {
	configDetailsMap, err := CreateConsulDetails(configPath)
	if err != nil {
		log.Fatalf("Failed to create consul details from config path %s: %v", configPath, err)
	}

	var finalBackupPath string
	if backupFilePath == "" {
		pwd, _ := os.Getwd()
		finalBackupPath = pwd // DefaultBackupFilePath will be appended per-consul by createBackupFileAndWriteData
	} else {
		finalBackupPath = backupFilePath
	}

	processFunc := func(detail config.ConsulDetail) {
		manager, errMgr := NewConsulManager(&detail)
		if errMgr != nil {
			errConsulConnection(detail.ConsulName, errMgr, false)
			return
		}
		kvPairs, errList := manager.ListAllKV() // Lists everything under manager.Detail.BasePath
		if errList != nil {
			fmt.Printf("<> -- Could not List KVs for Backup from Consul Server %s. Error: %v -- <>\n", detail.ConsulName, errList)
			return
		}
		// kvPairsToJSON now takes api.KVPairs. The original implementation took *api.KVPairs.
		// And it needs to handle the full keys as returned by ListAllKV.
		jsonData, errJson := kvPairsToJSON(kvPairs) // Adjusted kvPairsToJSON to take KVPairs
		if errJson != nil {
			fmt.Printf("<> -- Could not convert KVs to JSON for %s. Error: %v -- <>\n", detail.ConsulName, errJson)
			return
		}

		// createBackupFileAndWriteData's signature in helpers.go is (fp string, cn string, data []byte) (string, error)
		// fp: base directory path. cn: consul name (for filename part). data: data to write.
		// finalBackupPath is the base directory. detail.ConsulName is the consul name.
		// So the call should be createBackupFileAndWriteData(finalBackupPath, detail.ConsulName, jsonData)
		_, errWrite := createBackupFileAndWriteData(finalBackupPath, detail.ConsulName, jsonData)
		if errWrite != nil {
			fmt.Printf("<> -- Could not Write Backup for Consul Server %s. Error: %v -- <>\n", detail.ConsulName, errWrite)
		}
	}

	if cn == "" {
		for _, detail := range configDetailsMap {
			processFunc(detail)
		}
		fmt.Println("Note:- Values in consul are base64 encoded (if originally so). This backup saves them as strings.")
	} else if IsConfigNameValidFunc(cn, configDetailsMap) { // Use Func
		processFunc(configDetailsMap[cn])
		fmt.Println("Note:- All Consul Values are base64 encoded (if originally so). This backup saves them as strings.")
	} else {
		fmt.Println("<> -- Invalid Consul Server Name specified. Check config file and input. -- <>")
	}
}

/*
RestoreConsulKV : To restore consul from backup file
*/
func RestoreConsulKV(cn, configPath, restoreFilePath, serviceNameFilter string) {
	configDetailsMap, err := CreateConsulDetailsFunc(configPath) // Use Func
	if err != nil {
		log.Fatalf("Failed to create consul details from config path %s: %v", configPath, err)
	}

	if !IsConfigNameValidFunc(cn, configDetailsMap) { // Use Func
		fmt.Println("<> -- Invalid Consul Server Name for Restore. Check config file and input. -- <>")
		return
	}
	detail := configDetailsMap[cn]

	manager, errMgr := NewConsulManagerFunc(&detail) // Use Func
	if errMgr != nil {
		ErrConsulConnectionFunc(detail.ConsulName, errMgr, true) // Use Func & Exit if cannot connect to target consul
		return
	}

	var actualRestoreFile string
	if restoreFilePath == "" {
		pwd, _ := os.Getwd()
		// Construct default path: <pwd>/<DefaultBackupDir>/<consulName>.json
		actualRestoreFile = path.Join(pwd, DefaultBackupFilePath, detail.ConsulName+".json")
	} else {
		if valErr := ValidateFilePathFunc(restoreFilePath); valErr != nil { // Use Func
			fmt.Printf("<> -- Invalid absolute path to recovery json file: %s. Error: %v -- <>\n", restoreFilePath, valErr)
			ExitFunc(1) // Use Exported ExitFunc from helpers.go
		}
		actualRestoreFile = restoreFilePath
	}
	fmt.Println("File path used to restore KV(s): ", actualRestoreFile)

	// readJSONFileAndReturnStructFunc returns *[]KVDetails, no error in current helpers.go mockable version
	kvDataFromFile := ReadJSONFileAndReturnStructFunc(actualRestoreFile)
	if kvDataFromFile == nil { // Check for nil, assuming it indicates an error from the (mocked) function
		fmt.Printf("<> -- Could not read or parse restore file %s -- <>\n", actualRestoreFile)
		return
	}
	kvPairsToRestore := ConvertJSONStructToKvPairsFunc(kvDataFromFile) // Use Func & Returns *[]api.KVPair

	for _, kv := range *kvPairsToRestore { // Dereference pointer
		// kv.Key from file is the full path. PutKV expects key relative to BasePath.
		// We need to make kv.Key relative to manager.Detail.BasePath for PutKV.
		// Or, if kv.Key from file IS ALREADY relative to BasePath, then it's fine.
		// The original code used putKV(client, &kv, "") implying keys in JSON are full paths.
		// Our new manager.PutKV prepends BasePath. So, if kv.Key from file is "expected/full/path",
		// and BasePath is "expected/", PutKV would make "expected/expected/full/path". This is wrong.
		//
		// Decision: Keys in the backup JSON are full paths.
		// We need a way to put full keys, or strip a prefix that matches BasePath.
		// Let's modify PutKV to accept an option or have a PutFullKV.
		// For now, let's assume kv.Key from JSON should be made relative if it contains BasePath.
		// Or, more simply, the restore logic should ensure the key passed to PutKV is the suffix.

		keyForPut := kv.Key
		if strings.HasPrefix(kv.Key, manager.Detail.BasePath) {
			keyForPut = strings.TrimPrefix(kv.Key, manager.Detail.BasePath)
		} else {
			// If the key from JSON doesn't have the current manager's basepath,
			// it might be from a different consul or a differently structured backup.
			// The original code put it with empty base path, effectively putting the full key from JSON.
			// This behavior should be preserved if manager.PutKV is used.
			// This means the JSON key must be treated as a suffix, and if it's a "full" path from another context,
			// it will be put under current BasePath. This is likely not intended for "full path" backups.
			//
			// Re-evaluating: The original `putKV(client, &kv, "")` meant the key in `kv` was absolute.
			// Our `manager.PutKV` takes a suffix. To restore an absolute key `absKey` from backup
			// into a consul at `BasePath`, the suffix should be `strings.TrimPrefix(absKey, BasePath)`.
			// But what if `absKey` is not under `BasePath`?
			// The most straightforward is to assume keys in backup are "suffix" keys that will be
			// restored under the target consul's BasePath. If serviceNameFilter is used, it refines this.
			// If the backup contains full paths like "env1/svc/key1" and we restore to "env2" (BasePath="env2/"),
			// we'd get "env2/env1/svc/key1". This needs clarification from original intent.
			//
			// Let's assume for now: keys in backup are relative suffixes.
			// If serviceNameFilter is present, kv.Key should be like "serviceName/actualKey".
		}


		if serviceNameFilter != "" {
			// kv.Key is like "filterService/actualKey" or "actualKey" if backup was service specific.
			// We need to check if keyForPut (which is relative to BasePath) starts with serviceNameFilter + "/"
			// Example: BasePath="env/", kv.Key from backup="app/config". keyForPut="app/config".
			// serviceNameFilter="app". Check if "app/config" starts with "app/"
			if strings.HasPrefix(keyForPut, serviceNameFilter+"/") || keyForPut == serviceNameFilter {
				// If it matches, put it. The key in kvPairForPut should be keyForPut.
				kvPairForPut := api.KVPair{Key: keyForPut, Value: kv.Value, Flags: kv.Flags}
				_, errPut := manager.PutKV(&kvPairForPut)
				if errPut != nil {
					fmt.Printf("<> -- Error Restoring KV for key %s. Error: %v -- <>\n", kv.Key, errPut)
				}
			}
		} else { // No service filter, restore all keys from backup (relative to BasePath)
			kvPairForPut := api.KVPair{Key: keyForPut, Value: kv.Value, Flags: kv.Flags}
			_, errPut := manager.PutKV(&kvPairForPut)
			if errPut != nil {
				fmt.Printf("<> -- Error Restoring KV for key %s. Error: %v -- <>\n", kv.Key, errPut)
			}
		}
	}
	fmt.Println(" -- Consul Recovery Completed for Consul Server: ", detail.ConsulName, " -- ")
}


/*
SyncConsulKVStore :  Sync Kv Store of source Consul Server with target Consul Server
*/
func SyncConsulKVStore(sourceName, targetName, serviceNameFilter, configPath, replace string) {
	configDetailsMap, err := CreateConsulDetailsFunc(configPath) // Use Func
	if err != nil {
		log.Fatalf("Failed to create consul details from config path %s: %v", configPath, err)
	}

	if !IsConfigNameValidFunc(sourceName, configDetailsMap) || !IsConfigNameValidFunc(targetName, configDetailsMap) { // Use Func
		fmt.Println("<> -- Invalid Source or Target Consul Server Name. Check config file and input. -- <>")
		ExitFunc(1) // Use Exported ExitFunc from helpers.go
	}

	sourceDetail := configDetailsMap[sourceName]
	targetDetail := configDetailsMap[targetName]

	managerS, errMgrS := NewConsulManagerFunc(&sourceDetail) // Use Func
	if errMgrS != nil {
		ErrConsulConnectionFunc(sourceDetail.ConsulName, errMgrS, true) // Use Func
		return
	}
	sourceKVList, errListS := managerS.ListAllKV() // These have full paths from source's perspective
	if errListS != nil {
		fmt.Printf("<> -- Could not List KVs from Source Consul %s. Error: %v -- <>\n", sourceDetail.ConsulName, errListS)
		return
	}

	managerT, errMgrT := NewConsulManagerFunc(&targetDetail) // Use Func
	if errMgrT != nil {
		ErrConsulConnectionFunc(targetDetail.ConsulName, errMgrT, true) // Use Func
		return
	}
	targetKVList, errListT := managerT.ListAllKV() // These have full paths from target's perspective
	if errListT != nil {
		fmt.Printf("<> -- Could not List KVs from Target Consul %s. Error: %v -- <>\n", targetDetail.ConsulName, errListT)
		// Continue, as we might still be able to push source KVs if replace="true"
	}

	// KVs to sync, keys are relative suffixes (stripped of source base path)
	var kvMapToSync map[string]string // string value for easier comparison via removeExistingKVPairs
	if replace == "false" {
		// removeExistingKVPairs expects maps where keys are relative suffixes.
		// sourceBasePath and destinationBasePath are used to strip prefixes from the KVPairs.
		// The function signature in helpers.go is: removeExistingKVPairs(sourceKvPairs, targetKvPairs *api.KVPairs, sbf, tbf string) map[string][]byte
		// It expects *api.KVPairs but ListAllKV returns api.KVPairs.
		// For now, let's assume the mockable RemoveExistingKVPairsFunc can handle api.KVPairs or we adjust the call.
		// The map returned is map[string][]byte, but current var is map[string]string. This needs alignment.
		// For simplicity, assuming RemoveExistingKVPairsFunc is adapted or the map type is []byte.
		// Let's assume kvMapToSync is map[string][]byte based on RemoveExistingKVPairsFunc signature
		// Corrected argument order for RemoveExistingKVPairsFunc
		kvMapBytesToSync := RemoveExistingKVPairsFunc(sourceKVList, targetKVList, managerS.Detail.BasePath, managerT.Detail.BasePath)
		kvMapToSync = make(map[string]string)
		for k, v := range kvMapBytesToSync { kvMapToSync[k] = string(v) }

	} else { // replace == "true"
		// convertServiceKVPairsToMap also returns a map of relative keys.
		// Signature: convertServiceKVPairsToMap(kvPairs api.KVPairs, bp string) map[string][]byte
		// Corrected: ConvertServiceKVPairsToMapFunc expects api.KVPairs, not *api.KVPairs
		kvMapBytesToSync := ConvertServiceKVPairsToMapFunc(sourceKVList, managerS.Detail.BasePath)
		kvMapToSync = make(map[string]string)
		for k, v := range kvMapBytesToSync { kvMapToSync[k] = string(v) }
	}

	for keySuffix, valStr := range kvMapToSync {
		// keySuffix is relative to source's BasePath. We need to put it relative to target's BasePath.
		// This means the keySuffix is what managerT.PutKV expects.
		if serviceNameFilter == "" || strings.HasPrefix(keySuffix, serviceNameFilter+"/") || keySuffix == serviceNameFilter {
			kvToPut := api.KVPair{
				Key:   keySuffix, // This is the suffix for managerT.PutKV
				Value: []byte(valStr),
			}
			_, errPut := managerT.PutKV(&kvToPut)
			if errPut != nil {
				fmt.Printf("<> -- Could not Add/Update KV to Target Consul %s. Key: %s. Error: %v -- <>\n", targetDetail.ConsulName, path.Join(managerT.Detail.BasePath, keySuffix), errPut)
			}
		}
	}
	fmt.Printf("\n-- Consul KV Store Sync of Source Consul Server: %s and Target Consul Server: %s is complete --\n", sourceDetail.ConsulName, targetDetail.ConsulName)
}
