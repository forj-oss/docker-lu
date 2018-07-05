package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
)

var app struct {
	username  string
	uid       int
	groupname string
	gid       int
	help      bool
}

func main() {
	flag.BoolVar(&app.help, "help", false, "Help on docker-lu")

	flag.Parse()

	if app.help || len(flag.Args()) == 0 {
		showHelp()
	}
	update()
}

const (
	passwd     = "/etc/passwd"
	passwdTmp  = "/etc/passwd.tmp"
	passwdBack = "/etc/passwd.backup"
	group      = "/etc/group"
	groupTmp   = "/etc/group.tmp"
	groupBack  = "/etc/group.backup"
	cgroup     = "/proc/self/cgroup"
)

func update() {
	if err := checkRights(); err != nil {
		log.Fatalf("Unable to update files. %s. Aborted", err)
	}
	updateParse()

	defer cleanup()

	if err := updatePasswd(); err != nil {
		log.Fatalf("Unable to update %s. %s. Aborted", passwd, err)
	}
	if err := updateGroup(); err != nil {
		log.Fatalf("Unable to update %s. %s. Aborted", group, err)
	}
	applyUpdates()
	fmt.Println("DONE")
}

func cleanup() {
	if info, err := os.Stat(passwdTmp); err == nil && !info.IsDir() {
		os.Remove(passwdTmp)
	}
	if info, err := os.Stat(groupTmp); err == nil && !info.IsDir() {
		os.Remove(groupTmp)
	}
}

func checkRights() error {
	// Must be root
	if os.Getuid() != 0 {
		return fmt.Errorf("docker-lu must be executed as root. Exiting")
	}
	// Must be inside a container
	if _, err := os.Stat(cgroup); err != nil {
		return fmt.Errorf("Unable to check %s. %s", cgroup, err)
	}
	var cgroupData []byte
	if d, err := ioutil.ReadFile(cgroup); err != nil {
		return fmt.Errorf("Unable to read %s. %s", cgroup, err)
	} else {
		cgroupData = d
	}

	if ok, _ := regexp.Match("[0-9]+:[a-z_]*:/docker/[0-9a-f]*", cgroupData); !ok {
		return fmt.Errorf("docker-lu must be executed inside a container")
	}
	return nil
}

func applyUpdates() {
	if info, err := os.Stat(passwdTmp); err == nil && !info.IsDir() {
		if info, err = os.Stat(passwdBack); err == nil && !info.IsDir() {
			os.Remove(passwdBack)
		}
		os.Rename(passwd, passwdBack)
		os.Rename(passwdTmp, passwd)
	}
	if info, err := os.Stat(groupTmp); err == nil && !info.IsDir() {
		if info, err = os.Stat(groupBack); err == nil && !info.IsDir() {
			os.Remove(groupBack)
		}
		os.Rename(group, groupBack)
		os.Rename(groupTmp, group)
	}
	fmt.Printf("Passwd and group updated for user %s(%s) with uid:%d and gid:%d\n", app.username, app.groupname, app.uid, app.gid)
}

func updatePasswd() error {
	passwdFileRead, err := os.Open(passwd)
	if err != nil {
		return fmt.Errorf("Unable to read %s. %s", passwd, err)
	}
	defer passwdFileRead.Close()

	passwdFileWrite, err := os.Create(passwdTmp)
	if err != nil {
		return fmt.Errorf("Unable to write %s. %s", passwdTmp, err)
	}
	defer passwdFileWrite.Close()

	updReg, _ := regexp.Compile("(" + app.username + ":x:)([0-9]*):([0-9]*):")

	found := false
	scan := bufio.NewScanner(passwdFileRead)
	for scan.Scan() {
		line := scan.Text()
		if updReg.MatchString(line) {
			newline := updReg.ReplaceAllString(line, "${1}"+strconv.Itoa(app.uid)+":"+strconv.Itoa(app.gid)+":")
			if line != newline {
				log.Printf("%s update to apply.", passwd)
				line = newline
			}
			found = true
		}
		passwdFileWrite.WriteString(line + "\n")
	}
	if !found {
		return fmt.Errorf("user %s not found", app.username)
	}
	return nil
}

func updateGroup() error {
	groupFileRead, err := os.Open(group)
	if err != nil {
		return fmt.Errorf("Unable to read %s. %s", passwd, err)
	}
	defer groupFileRead.Close()

	groupFileWrite, err := os.Create(groupTmp)
	if err != nil {
		return fmt.Errorf("Unable to write %s. %s", passwdTmp, err)
	}
	defer groupFileWrite.Close()

	updReg, _ := regexp.Compile("(" + app.groupname + ":x:)([0-9]*):")

	found := false
	scan := bufio.NewScanner(groupFileRead)
	for scan.Scan() {
		line := scan.Text()
		if updReg.MatchString(line) {
			newline := updReg.ReplaceAllString(line, "${1}"+strconv.Itoa(app.gid)+":")
			if line != newline {
				log.Printf("%s update to apply.", group)
				line = newline
			}
			found = true
		}
		groupFileWrite.WriteString(line + "\n")
	}
	if !found {
		return fmt.Errorf("group %s not found", app.groupname)
	}
	return nil
}

func updateParse() {
	for iCount, value := range flag.Args() {
		switch iCount {
		case 0:
			if ok, _ := regexp.MatchString("^[a-z_][a-z0-9_]{0,30}$", value); ok {
				app.username = value
			} else {
				log.Fatalf("Username: Must respect [a-z_][a-z0-9_]{0,30}")
			}
		case 1:
			if ok, _ := regexp.MatchString("^[0-9]{0,10}$", value); ok {
				if v, err := strconv.Atoi(value); err != nil {
					log.Fatalf("UID: '%s' is an invalid number. Must be between 0 and 2147483647", value)
				} else {
					app.uid = v
				}
			} else {
				log.Fatalf("UID: '%s' is an invalid number. Must be between 0 and 2147483647", value)
			}
		case 2:
			if ok, _ := regexp.MatchString("^[a-z_][a-z0-9_]{0,30}$", value); ok {
				app.groupname = value
			} else {
				log.Fatalf("Username: Must respect [a-z_][a-z0-9_]{0,30}")
			}
		case 3:
			if ok, _ := regexp.MatchString("^[0-9]+$", value); ok {
				if v, err := strconv.Atoi(value); err != nil {
					log.Fatalf("GID: '%s' is an invalid number. Must be between 0 and 2147483647", value)
				} else {
					app.gid = v
				}
			} else {
				log.Fatalf("UID: '%s' is an invalid number. Must be between 0 and 2147483647", value)
			}
		}
	}
}

func showHelp() {
	fmt.Print(`docker-lu - small GO program to adapt container files, /etc/passwd & /etc/group
usage is:
--help : This help page

Arguments:
docker-lu <username> <uid> <groupname> <gid>
where:
- username :  is an existing user from /etc/passwd. Must respect [a-z_][a-z0-9_]{0,30}.
- uid :       is the User ID. It must be in the range of 0-2147483647
- groupname : is an existing group name from /etc/group. Must respect [a-z_][a-z0-9_]{0,30}.
- gid:        is the group ID. It must be in the range of 0-2147483647

This program will change the <username> <groupname> ids with ids given <uid> <gid>
`)
	os.Exit(0)
}
