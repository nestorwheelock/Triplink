package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/mkideal/cli"
)

type delBackupT struct {
	cli.Helper
	BackupIPtables bool   `cli:"t,iptables" usage:"Update iptables" dft:"false"`
	BackupIPset    bool   `cli:"s,ipset" usage:"Update ipset" dft:"false"`
	All            bool   `cli:"a,all" usage:"Backup ipset and iptables" dft:"false"`
	Yes            bool   `cli:"y,yes" usage:"Don't confirm deletion" dft:"false"`
	ConfigName     string `cli:"C,config" usage:"Specify the config to use" dft:"config.json"`
	Verbose        int    `cli:"v,verbose" usage:"Specify how much logs should be displayed" dft:"0"`
}

var delBackupCMD = &cli.Command{
	Name:    "deletebackup",
	Aliases: []string{"db", "delbak", "delbackup", "deleteb", "dback", "delb"},
	Desc:    "delete backups from ipset(-s) and (iptables -t)",
	Argv:    func() interface{} { return new(delBackupT) },
	Fn: func(ctx *cli.Context) error {
		if os.Getuid() != 0 {
			fmt.Println("You need to be root!")
			return nil
		}
		argv := ctx.Argv().(*delBackupT)
		verboseLevel = argv.Verbose
		if argv.All {
			argv.BackupIPset = true
			argv.BackupIPtables = true
		}
		if !argv.BackupIPtables && !argv.BackupIPset {
			LogInfo("nothing to do")
			return nil
		}
		logStatus, configFile := createAndValidateConfigFile(argv.ConfigName)
		if logStatus < 0 {
			return errors.New("config not found")
		}
		delBackup(configFile, argv.BackupIPset, argv.BackupIPtables, argv.Yes)
		return nil
	},
}

func delBackup(configFile string, deleteIPset, deleteIPtables, ignoreConfirm bool) {
	configFolder, _ := path.Split(configFile)
	iptablesFile := configFolder + "iptables.bak"
	ipsetFile := configFolder + "ipset.bak"

	if !ignoreConfirm {
		whatToDelete := ""
		if deleteIPset && deleteIPtables {
			whatToDelete = "IPset & IPtables"
		} else if deleteIPset {
			whatToDelete = "IPset"
		} else if deleteIPtables {
			whatToDelete = "IPtables"
		}
		reader := bufio.NewReader(os.Stdin)

		if y, _ := confirmInput("Do you really want to delete "+whatToDelete+" backup [y/n] > ", reader); y {
			fmt.Println("Exiting")
			return
		}
	}

	if deleteIPtables {
		_, err := os.Stat(iptablesFile)
		if err == nil {
			err = os.Remove(iptablesFile)
			if err != nil {
				LogError("Can't delete backup file: " + iptablesFile)
			} else {
				LogInfo("Successfully deleted IPtables backup")
			}
		} else {
			LogInfo("No IPtables backup found. Skipping")
		}
	}

	if deleteIPset {
		_, err := os.Stat(ipsetFile)
		if err == nil {
			err = os.Remove(ipsetFile)
			if err == nil {
				LogInfo("Successfully deleted IPset backup")
			} else {
				LogError("Can't delete backup file: " + ipsetFile)
			}
		} else {
			LogInfo("No IPset backup found. Skipping")
		}

	}
}
