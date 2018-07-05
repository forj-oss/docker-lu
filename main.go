package main

import (
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

func update() {
	checkRights()
	defer cleanup()
	updateParse()
	if err := updatePasswd(); err != nil {
		log.Fatal("Unable to update /etc/passwd. %s", err)
	}
	if err := updateGroup(); err != nil {
		log.Fatal("Unable to update /etc/passwd. %s", err)
	}
	applyUpdates()
}

func cleanup() {
	
}

func checkRights() {
	// Must be root
	// Must be inside a container
}

func applyUpdates() {

}

func updatePasswd() error {
	ioutil.ReadFile("/etc/passwd")
	return nil
}

func updateGroup() error {
	ioutil.ReadFile("/etc/passwd")
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
