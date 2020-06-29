package src

/*
	@author: Aviral Nigam
	@github: https://github.com/iAviPro
	@date: 15 Jun, 2020
*/

import (
	"flag"
	"fmt"
	"os"
)

// ExecuteGoConsulKV : Driver code for goConsulKV
func ExecuteGoConsulKV() {
	var sn, cn, props, config, replace, fp, source, target string
	fmt.Println("::==> Welcome to github.com/iAviPro/goConsulKV <==::")
	add := flag.NewFlagSet("add", flag.ExitOnError)
	delete := flag.NewFlagSet("delete", flag.ExitOnError)
	backup := flag.NewFlagSet("backup", flag.ExitOnError)
	restore := flag.NewFlagSet("restore", flag.ExitOnError)
	sync := flag.NewFlagSet("sync", flag.ExitOnError)

	// add arguments
	add.StringVar(&sn, "s", "", "Define service name. Default is empty.")
	add.StringVar(&config, "config", "", "Define consul config yml file. Default is ./config/consulConfig.yml")
	add.StringVar(&cn, "n", "", "Define consul name as per config yml. Default is empty string, which updates all consuls in config yml")
	add.StringVar(&replace, "replace", "false", "['true' | 'false'] Replaces the Value if Key already exists. Default is false.")
	add.StringVar(&props, "p", "", "Define '|' separated properties. Default is empty string")

	// delete arguments
	delete.StringVar(&sn, "s", "", "Define service name. Default is empty.")
	delete.StringVar(&config, "config", "", "Define consul config yml file. Default is ./config/consulConfig.yml")
	delete.StringVar(&cn, "n", "", "Define consul name as per config yml. Default is empty string, which updates all consuls in config yml")
	delete.StringVar(&props, "p", "", "Define '|' separated properties. Default is empty string")

	// backup arguments
	backup.StringVar(&config, "config", "", "Define consul config yml file. Default is ./config/consulConfig.yml")
	backup.StringVar(&cn, "n", "", "Define consul name as per config yml. Default is empty string, which updates all consuls in config yml")
	backup.StringVar(&fp, "save", "", "Define absolute directory path (without trailing '/') to save the backup file, given consul name will be the json file name. Default is empty string, which backs-up at ./backup/${consul-name}.json")

	// restore arguments
	restore.StringVar(&config, "config", "", "Define consul config yml file. Default is ./config/consulConfig.yml")
	restore.StringVar(&sn, "s", "", "Define service name. Default is empty.")
	restore.StringVar(&fp, "file", "", "Define absolute file path to recovery json file. Default is empty string, which tries to restore from ./backup/${consul-name}.json")
	restore.StringVar(&cn, "n", "", "Define consul name as per config yml. Default is empty string")

	// sync arguments
	sync.StringVar(&config, "config", "", "Define consul config yml file. Default is ./config/consulConfig.yml")
	sync.StringVar(&sn, "s", "", "Define service name. Default is empty.")
	sync.StringVar(&source, "source", "", "Define source consul name as per config yml. Default is empty.")
	sync.StringVar(&target, "target", "", "Define target consul name as per config yml. Default is empty.")
	sync.StringVar(&replace, "replace", "false", "['true' | 'false'] Replaces the Value if Key(s) already exists. Default is false.")

	switch os.Args[1] {
	case "add":
		{
			add.Parse(os.Args[2:])
			if props == "" || sn == "" {
				fmt.Println("Missing critical arguments for 'add' command. Execution stopped. Please use -help for more details.")
				os.Exit(1)
			}
			if config == "" || cn == "" {
				fmt.Println("Missing arguments. Default values for those arguments will be used. Please use -help for more details")
			}
			AddKVToConsul(sn, cn, props, config, replace)
		}

	case "delete":
		{
			delete.Parse(os.Args[2:])
			if sn == "" || props == "" {
				fmt.Println("Missing critical arguments for 'delete' command. Execution stopped. Please use -help for more details.")
				os.Exit(1)
			}
			if config == "" || cn == "" {
				fmt.Println("Missing arguments. Default values for those arguments will be used. Please use -help for more details")
			}
			DeleteKVFromConsul(sn, cn, props, config)

		}

	case "backup":
		{
			backup.Parse(os.Args[2:])
			if config == "" || cn == "" || fp == "" {
				fmt.Println("Missing arguments. Default values for those arguments will be used. Please use -help for more details")
			}
			BackupConsulKV(cn, config, fp)
		}
	case "restore":
		{
			restore.Parse(os.Args[2:])
			if cn == "" {
				fmt.Println("Missing critical arguments for 'restore' command. Execution stopped. Please use -help for more details.")
				os.Exit(1)
			}
			if config == "" || sn == "" || fp == "" {
				fmt.Println("Missing arguments. Default values for those arguments will be used. Please use -help for more details")
			}
			RestoreConsulKV(cn, config, fp, sn)
		}
	case "sync":
		{
			sync.Parse(os.Args[2:])
			if source == "" || target == "" {
				fmt.Println("Missing critical arguments for 'sync' command. Execution stopped. Please use -help for more details.")
				os.Exit(1)
			}
			if config == "" || sn == "" {
				fmt.Println("Missing arguments. Default values for those arguments will be used. Please use -help for more details")
			}
			SyncConsulKVStore(source, target, sn, config, replace)
		}
	}
}
