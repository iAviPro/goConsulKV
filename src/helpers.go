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

// ProcessKVPairs : Convert KVPairs to json
func processKVPairs(kvPairs *api.KVPairs, cn, fp string) []byte {
	var jsonbackup []*KVDetails
	for _, kv := range *kvPairs {
		// skip folder creation in consul as its autocreated.
		if kv.Key[len(kv.Key)-1:] == "/" {
			continue
		} else if string(kv.Value) == "" || kv.Value == nil {
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
		if er := config.ValidateConfigPath(cfp); er != nil {
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
