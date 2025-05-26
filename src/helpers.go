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

	"github.com/ConsulScale/config" // Corrected import path
	"github.com/hashicorp/consul/api"
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
func kvPairsToJSON(kvPairs api.KVPairs) ([]byte, error) { // Changed signature: *api.KVPairs -> api.KVPairs, added error return
	var jsonbackup []*KVDetails
	// Handle nil kvPairs explicitly if necessary, though ranging over nil slice is safe (0 iterations)
	if kvPairs == nil {
		// Return null or empty array based on desired JSON representation for nil input
		return json.Marshal(nil) // Or json.Marshal([]KVDetails{}) for "[]"
	}
	for _, kv := range kvPairs { // Range over api.KVPairs directly
		// skip folder creation in consul as its autocreated.
		if kv.Key[len(kv.Key)-1:] == "/" {
			continue
		} else if kv.Value == nil {
			continue
		} else {
			kvDetails := KVDetails{
				Key:   kv.Key,
				Value: kv.Value, // Value is []byte, matches struct
				Flags: kv.Flags,
			}
			jsonbackup = append(jsonbackup, &kvDetails)
		}
	}
	b, err := json.MarshalIndent(jsonbackup, "", "    ")
	if err != nil {
		// fmt.Println("error:", err) // Avoid printing directly in library functions
		return nil, err // Propagate error
	}
	// fmt.Println(string(b)) // Avoid printing directly
	return b, nil
}

// createBackupFileAndWriteData write file with backup data
// fp is the base directory path, cn is the consul name (used as part of filename)
func createBackupFileAndWriteData(fp string, cn string, data []byte) (string, error) { // Added string return for filepath
	// Ensure the base directory path exists
	if err := os.MkdirAll(fp, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create backup directory %s: %w", fp, err)
	}
	filepath := fp + "/" + cn + ".json" // Original logic for joining path
	err := ioutil.WriteFile(filepath, data, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write backup file %s: %w", filepath, err)
	}
	return filepath, nil // Return the path of the created file and nil error
}

// CreateConsulDetailsFunc points to the original CreateConsulDetails function.
// Tests can override this variable to mock the behavior of CreateConsulDetails.
var CreateConsulDetailsFunc = CreateConsulDetails // Already done

// CreateKVPairsFunc points to the original createKVPairs function.
var CreateKVPairsFunc = createKVPairs

// IsConfigNameValidFunc points to the original isConfigNameValid function.
var IsConfigNameValidFunc = isConfigNameValid

// ErrConsulConnectionFunc points to the original errConsulConnection function.
var ErrConsulConnectionFunc = errConsulConnection

// ValidateFilePathFunc points to the original ValidateFilePath function.
var ValidateFilePathFunc = ValidateFilePath

// ReadJSONFileAndReturnStructFunc points to the original readJSONFileAndReturnStruct function.
var ReadJSONFileAndReturnStructFunc = readJSONFileAndReturnStruct

// ConvertJSONStructToKvPairsFunc points to the original convertJSONStructToKvPairs function.
var ConvertJSONStructToKvPairsFunc = convertJSONStructToKvPairs

// KVPairsToJSONFunc points to the original kvPairsToJSON function.
var KVPairsToJSONFunc = kvPairsToJSON // Already done by implication

// CreateBackupFileAndWriteDataFunc points to the original createBackupFileAndWriteData function.
var CreateBackupFileAndWriteDataFunc = createBackupFileAndWriteData // Already done by implication

// RemoveExistingKVPairsFunc points to the original removeExistingKVPairs function.
var RemoveExistingKVPairsFunc = removeExistingKVPairs

// ConvertServiceKVPairsToMapFunc points to the original convertServiceKVPairsToMap function.
var ConvertServiceKVPairsToMapFunc func(kvPairs *api.KVPairs, bp string) map[string][]byte = convertServiceKVPairsToMap

// RemoveExistingKVPairsFunc points to the original removeExistingKVPairs function.
var RemoveExistingKVPairsFunc func(sourceKvPairs, targetKvPairs *api.KVPairs, sbf, tbf string) map[string][]byte = removeExistingKVPairs

// ExitFunc can be reassigned in tests to mock os.Exit.
var ExitFunc = os.Exit

// CreateConsulDetails : Read console params and yml file to generate console details
func CreateConsulDetails(cfp string) (map[string]config.ConsulDetail, error) { // Added error return
	var parseFile string
	if cfp == "" {
		parseFile = config.DefaultPathToEnvConfigFile
	} else {
		if er := ValidateFilePath(cfp); er != nil {
			parseFile = config.DefaultPathToEnvConfigFile
			fmt.Printf("\nError in config file path: %s \n", cfp)
			fmt.Println(er)
			// os.Exit(1) // Replaced os.Exit with error return
			return nil, fmt.Errorf("error in config file path %s: %w", cfp, er)
		} else {
			parseFile = cfp
		}
	}
	allConfigs, err := config.ParseConfigFile(parseFile)
	if err != nil {
		fmt.Printf("\nError in reading config yml file on path: %s \n", parseFile) // Use parseFile here
		fmt.Println(err)
		// os.Exit(1) // Replaced os.Exit with error return
		return nil, fmt.Errorf("error reading config yml file on path %s: %w", parseFile, err)
	}
	confMap := config.GetConsulConfigMap(allConfigs)
	return confMap, nil // Return map and nil error
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
func convertServiceKVPairsToMap(kvPairs *api.KVPairs, bp string) map[string][]byte { // Changed to *api.KVPairs
	var kvMap = make(map[string][]byte)
	if kvPairs == nil {
		return kvMap
	}
	for _, kv := range *kvPairs { // Dereference kvPairs
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
func removeExistingKVPairs(sourceKvPairs, targetKvPairs *api.KVPairs, sbf, tbf string) map[string][]byte { // Changed to *api.KVPairs
	targetKvMap := ConvertServiceKVPairsToMapFunc(targetKvPairs, tbf) // Use Func for internal call consistency
	var res = make(map[string][]byte)

	if sourceKvPairs == nil {
		return res
	}
	for _, sourceKv := range *sourceKvPairs { // Dereference sourceKvPairs
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

func errConsulConnection(name string, er error, exit bool) {
	if er != nil {
		fmt.Println(" <> -- Could Not Connect To Consul Server: ", name, " -- <>")
		fmt.Println(er)
		if exit {
			ExitFunc(1) // Use exported ExitFunc
		}
	} else {
		fmt.Printf("\n -- Successfully Connected To Consul Server: %s -- \n", name)
	}
}
