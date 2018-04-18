package plugin

import (
	"github.com/StackPointCloud/oneandone-flex-volume/pkg/flex"
)

// Attach volume to the node
func (v *VolumePlugin) Attach(options string, node string) (*flex.DriverStatus, error) {
	opt, err := v.newOptions(options)
	if err != nil {
		return nil, err
	}

	storage, err := v.manager.GetBlockstorage(opt.StorageID)
	if err != nil {
		return nil, err
	}

	return &flex.DriverStatus{
		Status:     flex.StatusSuccess,
		DevicePath: storage.Name,
	}, nil
}

// Detach the volume from the node
func (v *VolumePlugin) Detach(device, node string) (*flex.DriverStatus, error) {
	// storage, err := v.manager.GetBlockstorageByName(device)
	// if err != nil {
	// 	return nil, err
	// }

	// d, err := v.manager.FindServerFromNodeName(node)
	// if err != nil {
	// 	return nil, err
	// }

	// if storage.Server == nil {
	// 	err := v.manager.RemoveBlockStorageServer(storage.Id, d.Id)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }

	return &flex.DriverStatus{
		Status: flex.StatusSuccess,
	}, nil
}

// WaitForAttach no need to implement since we wait at the Attach command
func (v *VolumePlugin) WaitForAttach(device string, options string) (*flex.DriverStatus, error) {
	r := &flex.DriverStatus{
		Status: flex.StatusNotSupported,
	}
	return r, nil
}

// IsAttached checks for the volume to be attached to the node
func (v *VolumePlugin) IsAttached(options string, node string) (*flex.DriverStatus, error) {
	return &flex.DriverStatus{
		Status:   flex.StatusSuccess,
		Attached: true,
	}, nil
}
