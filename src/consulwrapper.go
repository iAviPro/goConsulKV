package src

/*
	@author: Aviral Nigam
	@github: https://github.com/iAviPro
	@date: 15 Jun, 2020
*/

import (
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/consul/api"
)

/*
ConnectConsul : Connect to consul server
*/
func ConnectConsul(address, datacentre, token string) (*api.Client, error) {
	config := api.DefaultConfig()
	config.Address = address
	config.Datacenter = datacentre
	config.Token = token

	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// putKV : Put a single key and value in consul key-value store(WriteOptions set to nil)
func putKV(client *api.Client, kvPair *api.KVPair, basePath string) (bool, error) {
	kv := client.KV()
	kvPair.Key = basePath + kvPair.Key
	_, e := kv.Put(kvPair, nil)
	if e != nil {
		return false, e
	}
	fmt.Printf("Added KV: %s = %s \n", kvPair.Key, kvPair.Value)
	return true, nil
}

// getKV to get key values using client
func getKV(client *api.Client, key, basePath string) (*api.KVPair, error) {
	kv := client.KV()
	pair, _, err := kv.Get(basePath+key, nil)
	if err != nil {
		return nil, err
	}
	return pair, nil
}

// deleteKV : deletes a single key (WriteOptions set to nil)
func deleteKV(client *api.Client, key, basePath string) (bool, error) {
	kv := client.KV()
	key = basePath + key
	_, err := kv.Delete(key, nil)
	if err != nil {
		return false, err
	}
	fmt.Printf("Deleted Key: %s\n", key)
	return true, nil
}

// listAllKV : Returns the list of All the KV pairs of a given consul
func listAllKV(client *api.Client, basePath string) (*api.KVPairs, error) {
	kv := client.KV()
	kvPairs, _, err := kv.List(basePath, nil)
	if err != nil {
		fmt.Println("There was an error in ListAllKV() Operation")
		return nil, err
	}
	/* Debug: Print KV List */
	// for _, kv := range kvPairs {
	// 	fmt.Printf("%s = %s \n", kv.Key, kv.Value)
	// }
	return &kvPairs, nil
}

// AddKVToConsul : Allows only to add KV-pairs in multi-consul setup, if the KV is already existing it will not update KV
func AddKVToConsul(sn, cn, props, config, replace string) {
	configMap := CreateConsulDetails(config)
	kvPairs := createKVPairs(props, sn)
	if cn == "" {
		for name, conf := range configMap {
			client, e := ConnectConsul(conf.BaseURL, conf.DataCentre, conf.Token)
			errConsulConnection(name, e, false)
			var addfail = make(map[string][]byte)
			for _, kv := range kvPairs {
				if replace == "false" {
					// check if the key exists
					if pair, _ := getKV(client, kv.Key, conf.BasePath); pair == nil {
						// put the key if it does not exist
						_, er := putKV(client, &kv, conf.BasePath)
						if er != nil {
							fmt.Println("<> -- Could not Update to Consul Server: " + name + "Following KV =>")
							fmt.Printf("\n%s = %s", kv.Key, kv.Value)
							fmt.Println(er)
						}
					} else {
						// if the key exists and replace is false, then put it in failed map
						addfail[kv.Key] = kv.Value
					}
				} else {
					_, er := putKV(client, &kv, conf.BasePath)
					if er != nil {
						fmt.Println("<> -- Could not Add Key to Consul Server: " + name + ". Following KV was not added =>")
						fmt.Printf("\n%s = %s", kv.Key, kv.Value)
						fmt.Println(er)
					}
				}
			}
			// print all the kv pairs that were not added
			if len(addfail) > 0 {
				fmt.Printf("\n <> -- Failed to Add Following KVs for Consul Server: %s -- <> \n", name)
				for k := range addfail {
					fmt.Printf("Key Already Exists: %s \n", k)
				}
				fmt.Println("\nNote:- Use --replace 'true' in the run command to replace values of existing keys. Use --help for more info.")
			}
		}
	} else if isConfigNameValid(cn, configMap) {
		conf := configMap[cn]
		client, e := ConnectConsul(conf.BaseURL, conf.DataCentre, conf.Token)
		errConsulConnection(conf.ConsulName, e, true)
		for _, kv := range kvPairs {
			_, err := putKV(client, &kv, conf.BasePath)
			if err != nil {
				fmt.Println("<> -- Could not Update KV(s) in Consul Server: " + cn + " -- <>")
				fmt.Printf("\n%s = %s", kv.Key, kv.Value)
				fmt.Println(err)
			}
		}
	} else {
		fmt.Println("<> -- Invalid Consul Server. Consul Server Name should be similar to name in config yml -- <>")
	}

}

/*
DeleteKVFromConsul : Deletes given keys from consul server
*/
func DeleteKVFromConsul(sn, cn, props, config string) {
	configMap := CreateConsulDetails(config)
	kvPairs := createKVPairs(props, sn)
	// delete from all consuls
	if cn == "" {
		for name, conf := range configMap {
			client, e := ConnectConsul(conf.BaseURL, conf.DataCentre, conf.Token)
			errConsulConnection(conf.ConsulName, e, false)
			for _, kv := range kvPairs {
				_, er := deleteKV(client, kv.Key, conf.BasePath)
				if er != nil {
					fmt.Println("<> -- Could not Delete Key(s) from Consul Server: " + name + " -- <>")
					fmt.Printf("\n%s", conf.BasePath+kv.Key)
					fmt.Println(er)
				}

			}
		}
	} else if isConfigNameValid(cn, configMap) {
		// delete from single consul
		conf := configMap[cn]
		client, e := ConnectConsul(conf.BaseURL, conf.DataCentre, conf.Token)
		errConsulConnection(conf.ConsulName, e, true)
		for _, kv := range kvPairs {
			_, err := deleteKV(client, kv.Key, conf.BasePath)
			if err != nil {
				fmt.Println("<> -- Could not Delete Key(s) from Consul Server: " + cn + " -- <>")
				fmt.Printf("\n%s", conf.BasePath+kv.Key)
				fmt.Println(err)
			}
		}
	} else {
		fmt.Println("<> -- Invalid Consul Server. Consul Server Name should be similar to name in config yml -- <>")
	}

}

/*
BackupConsulKV : Function to take json backups of consul and save to a backup file
Saves json file of format {"key": "key1", "value": "value1", "flags": "flag0"}
*/
func BackupConsulKV(cn, cp, fp string) {
	configMap := CreateConsulDetails(cp)
	var path string
	if fp == "" {
		pwd, _ := os.Getwd()
		path = pwd + DefaultBackupFilePath
	} else {
		path = fp
	}
	if cn == "" {
		for name, conf := range configMap {
			client, e := ConnectConsul(conf.BaseURL, conf.DataCentre, conf.Token)
			errConsulConnection(name, e, false)
			kvPairs, er := listAllKV(client, conf.BasePath)
			if er != nil {
				fmt.Println("<> -- Could not Backup KV(s) from Consul Server: ", name, " -- <>")
				fmt.Println(er)
			}
			data := kvPairsToJSON(kvPairs, conf.ConsulName, path)
			err := createBackupFileAndWriteData(path, conf.ConsulName, data)
			if err != nil {
				fmt.Println(err)
			}
		}
		fmt.Println("Note:- Values in consul are base64 encoded")
	} else if isConfigNameValid(cn, configMap) {
		conf := configMap[cn]
		client, e := ConnectConsul(conf.BaseURL, conf.DataCentre, conf.Token)
		errConsulConnection(conf.ConsulName, e, true)
		kvPairs, er := listAllKV(client, conf.BasePath)
		if er != nil {
			fmt.Println("<> -- Could not Backup KV(s) from Consul Server: ", conf.ConsulName, " -- <>")
			fmt.Println(er)
		}
		data := kvPairsToJSON(kvPairs, conf.ConsulName, path)
		err := createBackupFileAndWriteData(path, conf.ConsulName, data)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("Note:- All Consul Values are base64 encoded.")
	} else {
		fmt.Println("<> -- Invalid Consul Server. Consul Server Name should be similar to name in config yml -- <>")
	}
}

/*
RestoreConsulKV : To restore consul from backup file
Reads from json file of format {"key": "key1", "value": "value1", "flags": "flag0"} .
BackupConsulKV saves the backup in same format.
*/
func RestoreConsulKV(cn, cp, fp, sn string) {
	var file string
	configMap := CreateConsulDetails(cp)
	client, e := ConnectConsul(configMap[cn].BaseURL, configMap[cn].DataCentre, configMap[cn].Token)
	errConsulConnection(configMap[cn].ConsulName, e, true)
	if fp == "" {
		pwd, _ := os.Getwd()
		// default backup path if file path is ""
		file = pwd + DefaultBackupFilePath + "/" + configMap[cn].ConsulName + ".json"
	} else if ValidateFilePath(fp) != nil {
		file = fp
		fmt.Println("<> -- Invalid absolute path to recovery json file: ", file, " -- <>")
		os.Exit(1)
	} else {
		file = fp
	}
	fmt.Println("File path used to restore KV(s) (based on the parameters given) : ", file)
	kvStruct := readJSONFileAndReturnStruct(file)
	kvPairs := convertJSONStructToKvPairs(kvStruct)
	for _, kv := range *kvPairs {
		// for recovering only a particular service KVs
		if sn != "" {
			url := configMap[cn].BasePath + sn + "/"
			if strings.Contains(kv.Key, url) {
				// base path is empty as json already had the basepath within key
				putKV(client, &kv, "")
			} else {
				continue
			}
		} else {
			// for recovery of all the consul KVs
			// base path is empty as json already had the basepath within key
			putKV(client, &kv, "")
		}
	}
	fmt.Println(" -- Consul Recovery Completed for Consul Server: ", configMap[cn].ConsulName, " -- ")
}

/*
SyncConsulKVStore :  Sync Kv Store of source Consul Server with target Consul Server
*/
func SyncConsulKVStore(source, target, sn, config, replace string) {
	configMap := CreateConsulDetails(config)
	if !isConfigNameValid(source, configMap) || !isConfigNameValid(target, configMap) {
		fmt.Println("<> -- Invalid Consul Server. Consul Server Name should be similar to name in config yml -- <>")
		os.Exit(1)
	}

	// source client
	clientS, er := ConnectConsul(configMap[source].BaseURL, configMap[source].DataCentre, configMap[source].Token)
	errConsulConnection(configMap[source].ConsulName, er, true)

	sourceKVList, _ := listAllKV(clientS, configMap[source].BasePath)

	// target client
	clientT, e := ConnectConsul(configMap[target].BaseURL, configMap[target].DataCentre, configMap[target].Token)
	errConsulConnection(configMap[target].ConsulName, e, true)

	targetKVList, _ := listAllKV(clientT, configMap[target].BasePath)

	var kvPairsToSync = make(map[string][]byte)
	// if replace is false then only Keys that are in source but not in target KV store will be added
	if replace == "false" {
		kvPairsToSync = removeExistingKVPairs(sourceKVList, targetKVList, configMap[source].BasePath, configMap[target].BasePath)
	} else {
		kvPairsToSync = convertServiceKVPairsToMap(sourceKVList, configMap[source].BasePath)
	}
	if sn == "" {
		// for all services
		for key, val := range kvPairsToSync {
			kv := api.KVPair{
				Key:   configMap[target].BasePath + key,
				Value: val,
			}
			_, er := putKV(clientT, &kv, "")
			if er != nil {
				fmt.Println("<> -- Could not Add KV(s) to Consul Server: " + configMap[target].ConsulName + " -- <>")
				fmt.Printf("\n%s = %s", kv.Key, kv.Value)
				fmt.Println(er)
			}
		}
	} else {
		// all keys containing the given service name
		targetPath := sn + "/"
		for key, val := range kvPairsToSync {
			if strings.Contains(key, targetPath) {
				kv := api.KVPair{
					Key:   configMap[target].BasePath + key,
					Value: val,
				}
				_, er := putKV(clientT, &kv, "")
				if er != nil {
					fmt.Println("<> -- Could not Add KV(s) to Consul Server: " + configMap[target].ConsulName + " -- <>")
					fmt.Printf("\n%s = %s", kv.Key, kv.Value)
					fmt.Println(er)
				}
			}
		}
	}
	fmt.Printf("\n-- Consul KV Store Sync of Source Consul Server: %s and Target Consul Server: %s is complete --\n", configMap[source].ConsulName, configMap[target].ConsulName)
}
