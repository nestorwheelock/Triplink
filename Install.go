package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mkideal/cli"
)

type installT struct {
	cli.Helper
	ConfigName string `cli:"C,config" usage:"Specify the config to use" dft:"config.json"`
}

var installCMD = &cli.Command{
	Name:    "install",
	Aliases: []string{"install"},
	Desc:    "Setup automatic reports/updates easily",
	Argv:    func() interface{} { return new(installT) },
	Fn: func(ctx *cli.Context) error {
		if os.Getuid() != 0 {
			fmt.Println("You need to be root!")
			return nil
		}
		argv := ctx.Argv().(*installT)
		reader := bufio.NewReader(os.Stdin)
		i, text := WaitForMessage("What kind of system do you want to setup?\n[t] Tripwire\n[i] iptables/set\n> ", reader)
		if i == -1 {
			return nil
		}
		if i == 1 {
			text = strings.ToLower(text)
			if text == "t" {
				setTripwire(reader, argv.ConfigName)
			} else if text == "i" {
				setIP(reader, argv.ConfigName)
			} else {
				fmt.Println("What? Didn't understand '" + text + "'. Type 't' or 'i'")
				return nil
			}
		} else {
			return nil
		}

		//Tripwire related options
		//1. Update iplist (only blocking) periodically
		//2. Reporting + blocking periodically
		//3. Only reporting

		//firewall related options
		//1. Backup iptables & ipset
		//2. Restore iptables & ipset
		return nil
	},
}

func setIP(reader *bufio.Reader, config string) {
	i, opt := WaitForMessage("Backup or Restore?\n[b] Backup\n"+
		"[r] Restore\n> ", reader)
	if i != 1 {
		return
	}
	opt = strings.ToLower(opt)
	sMode := ""
	if opt == "b" {
		sMode = "backup"
	} else if opt == "r" {
		sMode = "restore"
	} else {
		fmt.Println("What? Didn't understand '" + opt + "'. Type 't' or 'i'")
		return
	}
	i, opt = WaitForMessage("What to "+sMode+"?\n[1] IPset\n"+
		"[2] IPtables\n"+
		"[3] both\n> ", reader)
	if i != 1 {
		return
	}
	ex, err := os.Executable()
	_ = ex
	if err != nil {
		panic(err)
	}

	i, text := WaitForMessage("In which period do you want to run this action [min/@reboot]: ", reader)
	if i != 1 {
		fmt.Println("Abort")
		return
	}
	var sTime string
	if text != "@reboot" {
		in, err := strconv.Atoi(text)
		if err != nil {
			fmt.Println("Not an integer")
			return
		}
		if in < 0 || in > 59 {
			fmt.Println("Your range must be between 0 and 60")
			return
		}
		sTime = "every " + text + " minutes"
		if in <= 1 {
			sTime = sTime[:len(sTime)-1]
		}
	} else {
		sTime = "after boot"
	}

	addCMD := sMode
	var description string

	if opt == "1" {
		addCMD += " -s"
		description = sMode + " IPset " + sTime
	} else if opt == "2" {
		addCMD += " -t -s=false"
		description = sMode + " IPtables " + sTime

	} else if opt == "3" {
		addCMD += " -s -t"
		description = sMode + " IPset and IPtables " + sTime

	} else {
		return
	}
	if text == "@reboot" {
		crontabReboot(addCMD, ex, description)
	} else {
		crontabPeriodically(text, addCMD, ex, description)
	}
}

func setTripwire(reader *bufio.Reader, c string) {
	config := getConfigPathFromHome(c)
	if handleConfig(config) {
		return
	}
	i, opt := WaitForMessage("How should Tripwire act?\n"+
		"[1] FETCH and block IPs from server based on a filter\t(blocker)\n"+
		"[2] REPORT IPs ONLY \t\t\t\t\t(reporter)\n"+
		"[3] REPORT and FETCH IPs to block using a filter\t(reporter and blocker)\n> ", reader)

	if i != 1 {
		return
	}

	if opt != "1" && opt != "2" && opt != "3" {
		fmt.Println("What? Enter 1,2 or 3")
		return
	}

	if opt != "2" {
		if y, _ := confirmInput("Do you want to update the filter assigned to the selected config (\""+c+"\") [y/n] ", reader); y {
			createFilter(config)
		}
	}

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	i, text := WaitForMessage("In which period do you want to run this action [min/@reboot]: ", reader)
	if i != 1 {
		fmt.Println("Abort")
		return
	}
	var sTime string
	if text != "@reboot" {
		in, err := strconv.Atoi(text)
		if err != nil {
			fmt.Println("Not an integer")
			return
		}
		if in < 0 || in > 59 {
			fmt.Println("Your range must be between 0 and 60")
			return
		}
		sTime = "every " + text + " minutes"
		if in <= 1 {
			sTime = sTime[:len(sTime)-1]
		}
	} else {
		sTime = "after boot"
	}

	var addCMD string
	var description string
	if opt == "1" {
		addCMD = "u" + " -C=\"" + c + "\""
		description = "Fetch and block IPs from server " + sTime + " (using \"" + config + "\" as configuration)"
	} else if opt == "3" {
		addCMD = "r -u" + " -C=\"" + c + "\""
		description = "Report IPs using tripwire AND Fetch and block IPs " + sTime + " (using \"" + config + "\" as configuration)"
	} else if opt == "2" {
		addCMD = "r" + " -C=\"" + c + "\""
		description = "Report IPs only (No blocking) " + sTime + " (using \"" + config + "\" as configuration)"
	} else {
		return
	}
	if text == "@reboot" {
		crontabReboot(addCMD, ex, description)
	} else {
		crontabPeriodically(text, addCMD, ex, description)
	}
}

func crontabReboot(addCMD, file, description string) {
	crontab("@reboot "+file+" "+addCMD+" > /dev/null", description)
}

func crontabPeriodically(interval, addCMD, file, description string) {
	crontab("*/"+interval+" * * * * "+file+" "+addCMD+" > /dev/null", description)
}

func crontab(content, description string) {
	err := writeCrontab(content, description)
	if err != nil {
		fmt.Println("Error writing crontab: " + err.Error())
	} else {
		fmt.Println("Installed successfully")
	}
	_, err = runCommand(nil, "systemctl restart cron")
	if err != nil {
		fmt.Println("Error restarting cron!")
	} else {
		fmt.Println("Restarted cron successfully")
	}
}

func writeCrontab(cronCommand, description string) error {
	f, err := os.OpenFile("/var/spool/cron/crontabs/root", os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	f.WriteString("\n# " + description + "\n" + cronCommand + "\n")
	f.Close()
	return nil
}

func handleConfig(config string) bool {
	_, err := os.Stat(config)
	if err != nil {
		fmt.Println("Config not found. Create a config with 'twreporter cc'.")
		return true
	}
	return false
}
