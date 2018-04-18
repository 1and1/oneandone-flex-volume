package plugin

import (
	"encoding/json"
	"fmt"

	"github.com/1and1/oneandone-flex-volume/pkg/flex"
	"github.com/1and1/oneandone-flex-volume/pkg/oneandone/cloud"
)

// VolumePlugin is a 1&1 flex volume plugin
type VolumePlugin struct {
	manager *cloud.OneandoneManager
}

// oneandoneOptions from the flex plugin
type oneandoneOptions struct {
	ApiKey         string `json:"kubernetes.io/secret/apiKey"`
	FsType         string `json:"kubernetes.io/fsType"`
	PVorVolumeName string `json:"kubernetes.io/pvOrVolumeName"`
	RW             string `json:"kubernetes.io/readwrite"`
	StorageName    string `json:"storageName,omitempty"`
	StorageID      string `json:"storageID,omitempty"`
}

// NewOneandoneVolumePlugin creates a 1&1 flex plugin
func NewOneandoneVolumePlugin(m *cloud.OneandoneManager) flex.VolumePlugin {
	return &VolumePlugin{
		manager: m,
	}
}

// Init driver
func (v *VolumePlugin) Init() (*flex.DriverStatus, error) {
	return &flex.DriverStatus{
		Status:  flex.StatusSuccess,
		Message: "1and1 flex driver initialized",
		Capabilities: &flex.DriverCapabilities{
			Attach:         true,
			SELinuxRelabel: true,
		},
	}, nil
}

func (v *VolumePlugin) newOptions(options string) (*oneandoneOptions, error) {
	opts := &oneandoneOptions{}
	if err := json.Unmarshal([]byte(options), opts); err != nil {
		return nil, err
	}
	return opts, nil
}

// GetVolumeName Retrieves a unique volume name
func (v *VolumePlugin) GetVolumeName(options string) (*flex.DriverStatus, error) {
	opt, err := v.newOptions(options)
	if err != nil {
		return nil, err
	}

	if opt.StorageID == "" {
		return nil, fmt.Errorf("1&1 volume needs StorageID property at flex options")
	}

	r := &flex.DriverStatus{
		Status:     flex.StatusSuccess,
		VolumeName: opt.StorageName,
	}
	return r, nil
}

// Mount volume at the dir where pods will use it
func (v *VolumePlugin) Mount(mountdir string, options string) (*flex.DriverStatus, error) {
	r := &flex.DriverStatus{
		Status:  flex.StatusNotSupported,
		Message: "mount",
	}
	return r, nil
}

// Unmount the volume at mount directory
func (v *VolumePlugin) Unmount(mountdir string) (*flex.DriverStatus, error) {
	r := &flex.DriverStatus{
		Status:  flex.StatusNotSupported,
		Message: "unmount",
	}
	return r, nil
}
