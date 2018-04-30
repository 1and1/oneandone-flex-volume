/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cloud

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/1and1/oneandone-cloudserver-sdk-go"
	"github.com/1and1/oneandone-flex-volume/helper"
)

// OneandoneManager communicates with the 1&1 API
type OneandoneManager struct {
	client *oneandone.API
	region string
}

// NewOneandoneManager returns a 1&1 manager
func NewOneandoneManager(token string) (*OneandoneManager, error) {
	_, file, no, ok := runtime.Caller(0)

	if ok {
		helper.DebugFile(fmt.Sprintf("Init called from %s#%d\n", file, no))
	} else {
		helper.DebugFile("not ok")
	}

	if token == "" {
		return nil, errors.New("1and1 token is empty")
	}

	helper.DebugFile(fmt.Sprintf("Using token -> %s", token))

	client := oneandone.New(token, oneandone.BaseUrl)

	m := &OneandoneManager{
		client: client,
	}

	return m, nil
}

type Result struct {
	Blockdevices []struct {
		Mountpoint string `json:"mountpoint"`
		Name       string `json:"name"`
		Children   []struct {
			Mountpoint string `json:"mountpoint"`
			Name       string `json:"name"`
		} `json:"children,omitempty"`
	} `json:"blockdevices"`
}

// GetServer retrieves the server by ID
func (m *OneandoneManager) GetServer(serverID string) (*oneandone.Server, error) {
	server, err := m.client.GetServer(serverID)

	if err != nil {
		return nil, fmt.Errorf("error fetching server %s not found %s", serverID, err.Error())
	}
	return server, nil
}

// GetBlockstorage given an unique 1&1 identifier returns the block storage
func (m *OneandoneManager) GetBlockstorage(storageID string) (*oneandone.BlockStorage, error) {
	storage, err := m.client.GetBlockStorage(storageID)

	if err != nil {
		return nil, fmt.Errorf("error fetching 1and1 block storage %s %s", storageID, err.Error())
	}

	return storage, nil
}

// GetBlockstorageByName given a name identifier returns the block storage
func (m *OneandoneManager) GetBlockstorageByName(name string) (*oneandone.BlockStorage, error) {
	storages, err := m.client.ListBlockStorages()

	if err != nil {
		return nil, err
	}

	for _, s := range storages {
		if strings.Contains(name, s.Name) {
			return &s, nil
		}
	}

	return nil, fmt.Errorf("storage with name %q was not found", name)
}

// AssignStorageAndWait attaches volume to given server
// it will wait until the attach action is completed
func (m *OneandoneManager) AssignStorageAndWait(storageID string, serverID string) error {
	storage, err := m.client.AddBlockStorageServer(storageID, serverID)
	if err != nil {
		return fmt.Errorf("error occured while adding storage to the server id %s, storage id %s, error %s", serverID, storageID, err.Error())
	}

	err = m.client.WaitForState(storage, "POWERED_ON", 10, 30)
	if err != nil {
		return err
	}

	return nil
}

// GetDeviceName finds system name of the block storage
func (m *OneandoneManager) GetDeviceName() (string, error) {
	deviceBaseName := "/dev/%s"

	var stdOut, stdErr bytes.Buffer
	cmd := exec.Command("lsblk", "-o", "MOUNTPOINT,NAME", "-J")
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("Error: %s, %s", err.Error(), stdErr.String())
	}

	resultObj := &Result{}

	json.Unmarshal(stdOut.Bytes(), resultObj)

	for _, b := range resultObj.Blockdevices {
		if b.Mountpoint == "" && len(b.Children) == 0 {
			return fmt.Sprintf(deviceBaseName, b.Name), nil
		}
	}
	return "", err
}

// GetMountPoint finds a mount point of the block storage
func (m *OneandoneManager) GetMountPoint(storageID string) (string, error) {
	deviceBaseName := "/dev/%s"

	var validID = regexp.MustCompile(`pvc-[A-Za-z0-9-]*`)
	var volumeName = validID.FindString(storageID)

	var stdOut, stdErr bytes.Buffer
	cmd := exec.Command("lsblk", "-o", "MOUNTPOINT,NAME", "-J")
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("Error: %s, %s", err.Error(), stdErr.String())
	}

	resultObj := &Result{}

	json.Unmarshal(stdOut.Bytes(), resultObj)

	for _, b := range resultObj.Blockdevices {
		match := validID.FindString(b.Mountpoint)
		if match == volumeName {
			return fmt.Sprintf(deviceBaseName, b.Name), nil
		}
	}
	return "", err
}

// RemoveBlockStorageServer detaches a disk to given server
func (m *OneandoneManager) RemoveBlockStorageServer(storageID string, serverID string) error {

	storage, err := m.client.GetBlockStorage(storageID)
	if err != nil {
		return err
	}
	if storage.Server != nil {
		_, err := m.client.RemoveBlockStorageServer(storageID, serverID)

		if err != nil {
			helper.DebugFile(fmt.Sprintf("failed once, trying again"))
			time.Sleep(1 * time.Second)
			_, err := m.client.RemoveBlockStorageServer(storageID, serverID)
			if err != nil {
				helper.DebugFile(fmt.Sprintf("error while RemoveBlockStorageServer %s", err.Error()))
				return err
			}
		}
	}

	return nil

}

// FindServerFromNodeName retrieves the server given the kubernetes node name
// Droplet name and Node name should match.
// If not, we will try to match the name with private and public IP
func (m *OneandoneManager) FindServerFromNodeName(node string) (*oneandone.Server, error) {
	// try to find server with same name as the kubernetes node
	servers, err := m.client.ListServers()
	if err != nil {
		return nil, err
	}

	for _, server := range servers {
		for _, ip := range server.Ips {
			if ip.Ip == node {
				return &server, nil
			}
		}
	}
	return nil, fmt.Errorf("could not match node name to server name")
}
