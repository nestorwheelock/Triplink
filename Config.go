package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

//Config the global config struct
type Config struct {
	Host    string `json:"host"`
	LogFile string `json:"logfile"`
	Token   string `json:"token"`
}

func getConfPath(homeDir string) string {
	return homeDir + "/" + ".tripwirereporter/"
}

func getConfFile(confPath string) string {
	return confPath + "conf.json"
}

func readConfig(file string) *Config {
	dat, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	res := Config{}
	err = json.Unmarshal(dat, &res)
	if err != nil {
		panic(err)
	}
	return &res
}

func saveConfig(configFile string, config *Config) error {
	sConf, err := json.Marshal(config)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(configFile, []byte(string(sConf)), 0600)
	if err != nil {
		return err
	}
	return nil
}

func createAndValidateConfigFile(logfile string) (int, string) {
	_, err := os.Stat(logfile)
	if err != nil {
		fmt.Println("Logfile doesn't exists")
		return -1, ""
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Couldn't retrieve homeDir!")
		return -1, ""
	}
	confPath := getConfPath(homeDir)
	confFile := getConfFile(confPath)
	_, err = os.Stat(confPath)
	if err != nil {
		err = os.MkdirAll(confPath, os.ModePerm)
		if err != nil {
			fmt.Println("Couldn't create configpath")
			return -1, ""
		}
		_, err = os.Create(confFile)
		if err != nil {
			fmt.Println("Couldn't create configfile")
			return -1, ""
		}
	}
	confStat, err := os.Stat(confFile)
	if err != nil {
		_, err = os.Create(confFile)
		if err != nil {
			fmt.Println("Couldn't create configfile")
			return -1, ""
		}
	}
	confStat, err = os.Stat(confFile)
	if err != nil {
		fmt.Println("Couldn't create configfile")
		return -1, ""
	}
	if confStat.Size() == 0 {
		return 0, confFile
	}

	return 1, confFile
}
