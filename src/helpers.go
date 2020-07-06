package src

/*
	@author: Aviral Nigam
	@github: https://github.com/iAviPro
	@date: 15 Jun, 2020
*/

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/hashicorp/consul/api"
	"github.com/iAviPro/goConsulKV/config"
)

// KVDetails : Struct to marshal backup KV pair json
type KVDetails struct {
	Key   string `json:"key"`
	Value []byte `json:"value"`
	Flags uint64 `json:"flags"`
}

// DefaultBackupFilePath : If file path to save dump is not given, then default file path is used.
const DefaultBackupFilePath string = "/backup"

func cleanKVPairs(props string) map[string]string {
	allKVPairs := make(map[string]string)
	allProps := strings.Split(props, "|")
	for _, prop := range allProps {
		prop = strings.TrimSpace(prop)
		KeyVal := strings.Split(prop, "=")
		key := strings.TrimSpace(KeyVal[0])
		value := strings.TrimSpace(KeyVal[1])
		allKVPairs[key] = value
	}
	return allKVPairs
}

// CreateKVPairs : Create KV Pairs
func createKVPairs(props, sName string) []api.KVPair {
	var allKVPairs []api.KVPair
	allProps := cleanKVPairs(props)
	for key, value := range allProps {
		pair := api.KVPair{
			Key:   sName + "/" + key,
			Value: []byte(value),
		}
		allKVPairs = append(allKVPairs, pair)
	}
	return allKVPairs
}

// kvPairsToJSON : Convert KVPairs to json
func kvPairsToJSON(kvPairs *api.KVPairs, cn, fp string) []byte {
	var jsonbackup []*KVDetails
	for _, kv := range *kvPairs {
		// skip folder creation in consul as its autocreated.
		if kv.Key[len(kv.Key)-1:] == "/" {
			continue
		} else if kv.Value == nil {
			continue
		} else {
			kvDetails := KVDetails{
				Key:   kv.Key,
				Value: kv.Value,
				Flags: kv.Flags,
			}
			jsonbackup = append(jsonbackup, &kvDetails)
		}
	}
	b, err := json.MarshalIndent(jsonbackup, "", "    ")
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Println(string(b))
	return b
}

// createBackupFileAndWriteData write file with backup data
func createBackupFileAndWriteData(fp, cn string, data []byte) error {
	filepath := fp + "/" + cn + ".json"
	os.MkdirAll(fp, os.ModePerm)
	err := ioutil.WriteFile(filepath, data, 0644)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

// CreateConsulDetails : Read console params and yml file to generate console details
func CreateConsulDetails(cfp string) map[string]config.ConsulDetail {
	var parseFile string
	if cfp == "" {
		parseFile = config.DefaultPathToEnvConfigFile
	} else {
		if er := ValidateFilePath(cfp); er != nil {
			parseFile = config.DefaultPathToEnvConfigFile
			fmt.Printf("\nError in config file path: %s \n", cfp)
			fmt.Println(er)
			os.Exit(1)
		} else {
			parseFile = cfp
		}
	}
	allConfigs, err := config.ParseConfigFile(parseFile)
	if err != nil {
		fmt.Printf("\nError in reading config yml file on path: %s \n", cfp)
		fmt.Println(err)
		os.Exit(1)
	}
	confMap := config.GetConsulConfigMap(allConfigs)
	return confMap
}

// ValidateFilePath : Validates if the path provide is a file
func ValidateFilePath(path string) error {
	s, err := os.Stat(path)
	if err != nil {
		return err
	}
	if s.IsDir() {
		return fmt.Errorf("'%s' is a directory, not a normal file", path)
	}
	return nil
}

func readJSONFileAndReturnStruct(fp string) *[]KVDetails {
	file, _ := ioutil.ReadFile(fp)
	var data []KVDetails
	_ = json.Unmarshal([]byte(file), &data)
	return &data
}

func convertJSONStructToKvPairs(kvDetails *[]KVDetails) *[]api.KVPair {
	var kvPairs []api.KVPair
	for _, data := range *kvDetails {
		var kv api.KVPair
		kv.Key = data.Key
		kv.Value = data.Value
		kv.Flags = data.Flags
		kvPairs = append(kvPairs, kv)
	}
	return &kvPairs
}

// convertServiceKVPairsToMap : key = only service path (removing the base path) and value is the KV value.
func convertServiceKVPairsToMap(kvPairs *api.KVPairs, bp string) map[string][]byte {
	var kvMap = make(map[string][]byte)
	for _, kv := range *kvPairs {
		// remove folders or keys with nil Kv Values
		if kv.Key[len(kv.Key)-1:] == "/" {
			continue
		} else if kv.Value == nil {
			continue
		} else {
			/*
				removing base path removes the conditional requirement of mapping the same base path
				in every consul. Each consul can have separate base path.
			*/
			k := removeBasePath(kv.Key, bp)
			kvMap[k] = kv.Value
		}
	}
	return kvMap
}

// removeExistingKVPairs : find keys of source that do not exists in target and return those kvPairs
func removeExistingKVPairs(sourceKvPairs, targetKvPairs *api.KVPairs, sbf, tbf string) map[string][]byte {
	targetKvMap := convertServiceKVPairsToMap(targetKvPairs, tbf)
	var res = make(map[string][]byte)

	for _, sourceKv := range *sourceKvPairs {
		if sourceKv.Key[len(sourceKv.Key)-1:] == "/" {
			continue
		} else if sourceKv.Value == nil {
			continue
		} else {
			// if value already exists in the targetKVStore then does not update
			k := removeBasePath(sourceKv.Key, sbf)
			if _, ok := targetKvMap[k]; !ok {
				res[k] = sourceKv.Value
			}
		}
	}
	return res
}

func isConfigNameValid(name string, consulData map[string]config.ConsulDetail) bool {
	if _, ok := consulData[name]; ok {
		return true
	}
	return false
}

func removeBasePath(key, bp string) string {
	k := strings.Replace(key, bp, "", 1)
	return k
}
