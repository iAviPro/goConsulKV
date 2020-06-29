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

// ConnectConsul : Connect to consul server of given details
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

// PutKV : Put a single key and value in consul key-value store(WriteOptions set to nil)
func PutKV(client *api.Client, kvPair *api.KVPair, basePath string) (bool, error) {
	kv := client.KV()
	kvPair.Key = basePath + kvPair.Key
	_, e := kv.Put(kvPair, nil)
	if e != nil {
		return false, e
	}
	fmt.Printf("Added KV: %s = %s \n", kvPair.Key, kvPair.Value)
	return true, nil
}

// GetKV to get key values using client
func GetKV(client *api.Client, key, basePath string) (*api.KVPair, error) {
	kv := client.KV()
	pair, _, err := kv.Get(basePath+key, nil)
	if err != nil {
		return nil, err
	}
	return pair, nil
}

// DeleteKV : deletes a single key (WriteOptions set to nil)
func DeleteKV(client *api.Client, key, basePath string) (bool, error) {
	kv := client.KV()
	key = basePath + key
	_, err := kv.Delete(key, nil)
	if err != nil {
		return false, err
	}
	fmt.Printf("Deleted Key: %s\n", key)
	return true, nil
}

// ListAllKV : Returns the list of All the KV pairs of a given consul
func ListAllKV(client *api.Client, basePath string) (*api.KVPairs, error) {
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
			if e != nil {
				fmt.Println("Could not Connect to Consul Name: " + name)
				fmt.Println(e)
			} else {
				fmt.Printf("\n-- Connected to Consul Name: %s --\n", name)
			}
			var addfail = make(map[string][]byte)
			for _, kv := range kvPairs {
				if replace == "false" {
					// check if the key exists
					if pair, _ := GetKV(client, kv.Key, conf.BasePath); pair == nil {
						// put the key if it does not exist
						_, er := PutKV(client, &kv, conf.BasePath)
						if er != nil {
							fmt.Println("<> -- Could not Update to Consul Name: " + name + "Following KV =>")
							fmt.Printf("\n%s = %s", kv.Key, kv.Value)
							fmt.Println(er)
						}
					} else {
						// if the key exists and replace is false, then put it in failed map
						addfail[kv.Key] = kv.Value
					}
				} else {
					_, er := PutKV(client, &kv, conf.BasePath)
					if er != nil {
						fmt.Println("<> -- Could not Add Key to Consul Name: " + name + ". Following KV was not added =>")
						fmt.Printf("\n%s = %s", kv.Key, kv.Value)
						fmt.Println(er)
					}
				}
			}
			// print all the kv pairs that were not added
			if len(addfail) > 0 {
				fmt.Printf("\n <> -- Failed to Add Following KVs for Consul Name: %s -- <> \n", name)
				for k := range addfail {
					fmt.Printf("Key Already Exists: %s \n", k)
				}
				fmt.Println("\nNote:- Use --replace 'true' in the run command to replace values of existing keys. Use --help for more info.")
			}
		}
	} else {
		conf := configMap[cn]
		client, e := ConnectConsul(conf.BaseURL, conf.DataCentre, conf.Token)
		if e != nil {
			fmt.Println("Could not Connect to Consul Name: " + conf.ConsulName)
			fmt.Println(e)
		} else {
			fmt.Printf("\n-- Connected to Consul Name: %s --\n", conf.ConsulName)
		}
		for _, kv := range kvPairs {
			_, err := PutKV(client, &kv, conf.BasePath)
			if err != nil {
				fmt.Println("<> -- Could not Update to Consul Name: " + cn + "Following KV =>")
				fmt.Printf("\n%s = %s", kv.Key, kv.Value)
				fmt.Println(err)
			}
		}
	}

}

// DeleteKVFromConsul : Deletes given prop keys from consul
func DeleteKVFromConsul(sn, cn, props, config string) {
	configMap := CreateConsulDetails(config)
	kvPairs := createKVPairs(props, sn)
	if cn == "" {
		for name, conf := range configMap {
			client, e := ConnectConsul(conf.BaseURL, conf.DataCentre, conf.Token)
			if e != nil {
				fmt.Println("Could not Connect to Consul Name: " + name)
				fmt.Println(e)
			} else {
				fmt.Printf("\n-- Connected to Consul Name: %s --\n", name)
			}
			for _, kv := range kvPairs {
				_, er := DeleteKV(client, kv.Key, conf.BasePath)
				if er != nil {
					fmt.Println("<> -- Could not Delete Key from Consul Name: " + name + ". Following KV was not deleted =>")
					fmt.Printf("\n%s", conf.BasePath+kv.Key)
					fmt.Println(er)
				}

			}
		}
	} else {
		conf := configMap[cn]
		client, e := ConnectConsul(conf.BaseURL, conf.DataCentre, conf.Token)
		if e != nil {
			fmt.Println("Could not Connect to Consul Name: " + conf.ConsulName)
			fmt.Println(e)
		} else {
			fmt.Printf("\n-- Connected to Consul Name: %s --\n", conf.ConsulName)
		}
		for _, kv := range kvPairs {
			_, err := DeleteKV(client, kv.Key, conf.BasePath)
			if err != nil {
				fmt.Println("<> -- Could not Delete from Consul Name: " + cn + "Following Key(s) =>")
				fmt.Printf("\n%s", conf.BasePath+kv.Key)
				fmt.Println(err)
			}
		}
	}

}

// BackupConsulKV : Function to take json backups of consul and save to a backup file
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
			if e != nil {
				fmt.Println("Could not Connect to Consul Name: " + name)
				fmt.Println(e)
			} else {
				fmt.Printf("\n-- Connected to Consul Name: %s --\n", name)
			}
			kvPairs, er := ListAllKV(client, conf.BasePath)
			if er != nil {
				fmt.Println("<> -- Could not Backup from Consul Name: " + name)
				fmt.Println(er)
			}
			data := processKVPairs(kvPairs, conf.ConsulName, path)
			err := createBackupFileAndWriteData(path, conf.ConsulName, data)
			if err != nil {
				fmt.Println(err)
			}
		}
		fmt.Println("Note:- Values in consul are base64 encoded")
	} else {
		conf := configMap[cn]
		client, e := ConnectConsul(conf.BaseURL, conf.DataCentre, conf.Token)
		if e != nil {
			fmt.Println("Could not Connect to Consul Name: " + conf.ConsulName)
			fmt.Println(e)
		} else {
			fmt.Printf("\n-- Connected to Consul Name: %s --\n", conf.ConsulName)
		}
		kvPairs, er := ListAllKV(client, conf.BasePath)
		if er != nil {
			fmt.Println("<> -- Could not Backup from Consul Name: " + conf.ConsulName)
			fmt.Println(er)
		}
		data := processKVPairs(kvPairs, conf.ConsulName, path)
		err := createBackupFileAndWriteData(path, conf.ConsulName, data)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("Note:- Values in consul are base64 encoded")
	}
}

// RestoreConsulKV : To restore consul from backup file
func RestoreConsulKV(cn, cp, fp, sn string) {
	var file string
	configMap := CreateConsulDetails(cp)
	client, e := ConnectConsul(configMap[cn].BaseURL, configMap[cn].DataCentre, configMap[cn].Token)
	if e != nil {
		fmt.Println("Could not Connect to Consul Name: " + configMap[cn].ConsulName)
		fmt.Println(e)
	} else {
		fmt.Printf("\n-- Connected to Consul Name: %s --\n", configMap[cn].ConsulName)
	}
	if fp == "" {
		pwd, _ := os.Getwd()
		// default backup path if file path is ""
		file = pwd + DefaultBackupFilePath + "/" + configMap[cn].ConsulName + ".json"
	} else if ValidateFilePath(fp) != nil {
		file = fp
		fmt.Println("Invalid absolute path to recovery json file. Given path: ", file)
		os.Exit(1)
	} else {
		file = fp
	}
	fmt.Println("File path for recovery used (based on the parameters given) : ", file)
	kvStruct := readJSONFileAndReturnStruct(file)
	kvPairs := convertJSONStructToKvPairs(kvStruct)
	for _, kv := range *kvPairs {
		// for recovering only a particular service KVs
		if sn != "" {
			url := configMap[cn].BasePath + sn + "/"
			if strings.Contains(kv.Key, url) {
				// base path is empty as json already had the basepath within key
				PutKV(client, &kv, "")
			} else {
				continue
			}
		} else {
			// for recovery of all the consul KVs
			// base path is empty as json already had the basepath within key
			PutKV(client, &kv, "")
		}
	}
	fmt.Println(" -- Consul Recovery Completed for name: ", configMap[cn].ConsulName)
}

// SyncConsulKVStore :  Sync Kv Store of source consul name with target consul name
func SyncConsulKVStore(source, target, sn, config, replace string) {
	configMap := CreateConsulDetails(config)
	if !isConfigNameValid(source, configMap) || !isConfigNameValid(target, configMap) {
		fmt.Println("Invalid Consul Name mentioned, consul name should be same as the config yml file")
		os.Exit(1)
	}

	// source client
	clientS, er := ConnectConsul(configMap[source].BaseURL, configMap[source].DataCentre, configMap[source].Token)
	if er != nil {
		fmt.Println("Could not Connect to Consul Name: " + configMap[source].ConsulName)
		fmt.Println(er)
	} else {
		fmt.Printf("\n-- Connected to Consul Name: %s --\n", configMap[source].ConsulName)
	}

	sourceKVList, _ := ListAllKV(clientS, configMap[source].BasePath)

	// target client
	clientT, e := ConnectConsul(configMap[target].BaseURL, configMap[target].DataCentre, configMap[target].Token)
	if e != nil {
		fmt.Println("Could not Connect to Consul Name: " + configMap[target].ConsulName)
		fmt.Println(e)
	} else {
		fmt.Printf("\n-- Connected to Consul Name: %s --\n", configMap[target].ConsulName)
	}
	targetKVList, _ := ListAllKV(clientT, configMap[target].BasePath)
	var kvPairsToSync = make(map[string][]byte)
	// if replace is false then only Keys that are in source but not in target KV store will be added
	if replace == "false" {
		kvPairsToSync = findUncommonKVPairs(sourceKVList, targetKVList)
	} else {
		kvPairsToSync = convertKVPairsToMap(sourceKVList)
	}
	if sn == "" {
		for key, val := range kvPairsToSync {
			kv := api.KVPair{
				Key:   key,
				Value: val,
			}
			_, er := PutKV(clientT, &kv, "")
			if er != nil {
				fmt.Println("<> -- Could not Add Key to Consul Name: " + configMap[target].ConsulName + ". Following KV was not added =>")
				fmt.Printf("\n%s = %s", kv.Key, kv.Value)
				fmt.Println(er)
			}
		}
	} else {
		targetPath := configMap[target].BasePath + sn + "/"
		for key, val := range kvPairsToSync {
			if strings.Contains(key, targetPath) {
				kv := api.KVPair{
					Key:   key,
					Value: val,
				}
				_, er := PutKV(clientT, &kv, "")
				if er != nil {
					fmt.Println("<> -- Could not Add Key to Consul Name: " + configMap[target].ConsulName + ". Following KV was not added =>")
					fmt.Printf("\n%s = %s", kv.Key, kv.Value)
					fmt.Println(er)
				}
			}
		}
	}
	fmt.Printf("\n-- Consul KV Store Sync of Source Consul: %s and Target Consul: %s is complete --\n", configMap[source].ConsulName, configMap[target].ConsulName)
}
