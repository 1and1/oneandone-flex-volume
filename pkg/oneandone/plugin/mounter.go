package plugin

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/1and1/oneandone-flex-volume/helper"
	"github.com/1and1/oneandone-flex-volume/pkg/flex"

	"golang.org/x/sys/unix"
)

// MountDevice mounts the volume as a device
func (v *VolumePlugin) MountDevice(mountdir, device string, options string) (*flex.DriverStatus, error) {
	helper.DebugFile(fmt.Sprintf("Device Name %s", device))

	opt, err := v.newOptions(options)
	if err != nil {
		return nil, err
	}

	serverID, err := helper.GetServerID()

	storage, err := v.manager.GetBlockstorage(opt.StorageID)
	if err != nil {
		return nil, err
	}

	if storage.Server == nil {
		err := v.manager.AssignStorageAndWait(storage.Id, serverID)
		if err != nil {
			helper.DebugFile(fmt.Sprintf("Error: %s", err.Error()))
			return nil, err
		}
	}
	storage, err = v.manager.GetBlockstorage(opt.StorageID)
	if err != nil {
		return nil, err
	}

	err = v.internalMount(mountdir, fmt.Sprintf("/dev/disk/by-id/scsi-3%s", storage.UUID), opt.FsType)
	if err != nil {
		return nil, err
	}

	return &flex.DriverStatus{
		Status: flex.StatusSuccess,
	}, nil
}

// UnmountDevice from the node
func (v *VolumePlugin) UnmountDevice(device string) (*flex.DriverStatus, error) {
	helper.DebugFile(fmt.Sprintf("Unmounting Device %s", device))

	storage, err := v.manager.GetBlockstorageByName(device)
	if err != nil {
		return nil, err
	}

	if err := v.internalUnmount(device); err != nil {
		helper.DebugFile(fmt.Sprintf("internalUnmount failure  %s", err.Error()))
		return nil, err
	}

	serverID, err := helper.GetServerID()

	if storage.Server != nil {
		err := v.manager.RemoveBlockStorageServer(storage.Id, serverID)
		if err != nil {
			helper.DebugFile(fmt.Sprintf("RemoveBlockStorageServer failure  %s", err.Error()))
			return nil, err
		}
	}

	r := &flex.DriverStatus{
		Status: flex.StatusSuccess,
	}
	return r, nil
}

func (v *VolumePlugin) isMounted(targetDir string) (bool, error) {
	findmntCmd := exec.Command("findmnt", "-n", targetDir)
	findmntStdout, err := findmntCmd.StdoutPipe()
	if err != nil {
		return false, fmt.Errorf("could not get findmount stdout pipe: %s", err.Error())
	}

	if err = findmntCmd.Start(); err != nil {
		return false, fmt.Errorf("findmnt failed to start: %s", err.Error())
	}

	findmntScanner := bufio.NewScanner(findmntStdout)
	findmntScanner.Split(bufio.ScanWords)
	findmntScanner.Scan()
	if findmntScanner.Err() != nil {
		return false, fmt.Errorf("could not get findmount output: %s", findmntScanner.Err().Error())
	}

	findmntText := findmntScanner.Text()
	if err = findmntCmd.Wait(); err != nil {
		_, isExitError := err.(*exec.ExitError)
		if !isExitError {
			return false, fmt.Errorf("findmount command failed: %s", err.Error())
		}
	}

	return findmntText == targetDir, nil
}

func (v *VolumePlugin) internalMount(targetDir string, device string, fsType string) error {
	helper.DebugFile("Mounting " + device)
	if fsType == "" {
		// default to ext4
		fsType = "ext4"
	}

	var res unix.Stat_t
	if err := unix.Stat(device, &res); err != nil {
		return fmt.Errorf("could not stat device %s: %s", device, err.Error())
	}

	if res.Mode&unix.S_IFMT != unix.S_IFBLK {
		return fmt.Errorf("device %s is not a block device", device)
	}

	mounted, err := v.isMounted(targetDir)
	if err != nil {
		return err
	}
	if mounted {
		return nil
	}

	format, err := v.currentFormat(device)
	if err != nil {
		return err
	}

	if !strings.Contains(format, fsType) {
		mkfsCmd := exec.Command("mkfs", "-t", fsType, device)
		if mkfsOut, err := mkfsCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("mkfs -t %s %s failed with error [%s] and output [%s]", fsType, device, err.Error(), string(mkfsOut))
		}
	}

	if _, err := os.Stat(targetDir); err != nil {
		if err := os.MkdirAll(targetDir, 0777); err != nil {
			return fmt.Errorf("could not create directory %s: %s", targetDir, err.Error())
		}
	}

	mountCmd := exec.Command("mount", device, targetDir)
	if mountOut, err := mountCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("mounting device %s at dir %s failed with error [%s] and output [%s] ", device, targetDir, err.Error(), string(mountOut))
	}

	return nil
}

func (v *VolumePlugin) internalUnmount(targetDir string) error {
	helper.DebugFile("targetDir: " + targetDir)
	mounted, err := v.isMounted(targetDir)
	if err != nil {
		return err
	}
	if !mounted {
		helper.DebugFile("!mounted")
		return nil
	}

	umountCmd := exec.Command("umount", targetDir)
	if umountOut, err := umountCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("!!!!unmounting the device at %s failed with error [%s] and output [%s]", targetDir, err.Error(), string(umountOut))
	}

	//Cleanup after unmount
	if os.RemoveAll(targetDir); err != nil {
		return fmt.Errorf("Error while trying to remove %s, %s", targetDir, err)
	}
	return nil
}

func (v *VolumePlugin) currentFormat(device string) (string, error) {

	lsblkCmd := exec.Command("lsblk", "-n", "-o", "FSTYPE", device)
	lsblkOut, err := lsblkCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("lsblk -n -o FSTYPE %s: output[%s] error[%s]", device, string(lsblkOut), err.Error())
	}

	output := strings.TrimSuffix(string(lsblkOut), "\n")
	lines := strings.Split(output, "\n")
	if lines[0] != "" {
		// The device is formatted
		return lines[0], nil
	}

	if len(lines) == 1 {
		// The device is unformatted and has no dependent devices
		return "", nil
	}

	// The device has dependent devices, most probably partitions (LVM, LUKS
	// and MD RAID are reported as FSTYPE and caught above).
	return "unknown data, probably partitions", nil
}
