package helper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

//DebugFile writes debug messages to /tmp/oneandone.log
func DebugFile(msg string) {
	isDebug := true

	if !isDebug {
		return
	}

	file := "/tmp/oneandone.log"
	var f *os.File
	t := time.Now()
	if _, err := os.Stat(file); os.IsNotExist(err) {
		f, err = os.Create(file)
		if err != nil {
			panic(err)
		}
	} else {
		f, err = os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			panic(err)
		}
	}
	defer f.Close()

	if _, err := f.WriteString(fmt.Sprintf("%s %s", t.Format(time.RFC822), msg+"\n")); err != nil {
		panic(err)
	}
}

//GetServerID gets server ID 1&1 Cloud Server Metadata API
func GetServerID() (string, error) {
	request, err := http.NewRequest("GET", "http://169.254.169.254/latest/meta_data/server_id", nil)
	if err != nil {
		return "", err
	}

	client := http.Client{}

	response, err := client.Do(request)
	if err != nil {
		return "", err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return strings.TrimRight(string(body), "\n"), nil
}

//Lsblk returns parsed results from `lsblk` command executed on the host
func Lsblk() (*Result, error) {
	cmd := exec.Command("lsblk", "-J", "-o", "NAME,MOUNTPOINT,TYPE,FSTYPE")

	data, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	result := &Result{}

	err = json.Unmarshal(data, result)
	if err != nil {
		return nil, err
	}

	return result, err
}

//Result struct representing `lsblk`` output
type Result struct {
	Devices []Device `json:"blockdevices"`
}

type Device struct {
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	Mountpoint string    `json:"mountpoint"`
	FSType     string    `json:"fstype"`
	Children   *[]Device `json:"children"`
}

//ResultDiff compares two Result structs and returns the differential.
func ResultDiff(oldV, newV Result) (toreturn []*Device) {
	var (
		// lenMin   int
		longest  Result
		shortest Result
		// f        *Device
		found bool
	)
	// Determine the shortest length and the longest slice
	if len(oldV.Devices) == 0 {
		toreturn = append(toreturn, &newV.Devices[len(newV.Devices)-1])
	} else if len(oldV.Devices) < len(newV.Devices) {
		longest = newV
		shortest = oldV

	} else {
		longest = oldV
		shortest = newV
	}

	diff := make(map[string]Device)
	for i, l := range longest.Devices {
		if l.Children == nil && l.Type != "rom" {
			for _, s := range shortest.Devices {
				if s.Children == nil && s.Type != "rom" {
					if len(shortest.Devices) > i {
						if l.Name == s.Name {
							found = true
							break
						} else {
							found = false
						}
					}
				}
			}
			if !found {
				diff[l.Name] = l
			}
		}
	}

	toreturn = make([]*Device, 0)
	for _, v := range diff {
		toreturn = append(toreturn, &v)
	}

	return toreturn
}
