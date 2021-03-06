package devicepathresolver

import (
	"strings"
	"time"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshsettings "github.com/cloudfoundry/bosh-agent/settings"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

type mappedDevicePathResolver struct {
	diskWaitTimeout time.Duration
	fs              boshsys.FileSystem
}

func NewMappedDevicePathResolver(
	diskWaitTimeout time.Duration,
	fs boshsys.FileSystem,
) mappedDevicePathResolver {
	return mappedDevicePathResolver{fs: fs, diskWaitTimeout: diskWaitTimeout}
}

func (dpr mappedDevicePathResolver) GetRealDevicePath(diskSettings boshsettings.DiskSettings) (string, bool, error) {
	stopAfter := time.Now().Add(dpr.diskWaitTimeout)

	devicePath := diskSettings.Path

	realPath, found := dpr.findPossibleDevice(devicePath)

	for !found {
		if time.Now().After(stopAfter) {
			return "", true, bosherr.Errorf("Timed out getting real device path for %s", devicePath)
		}

		time.Sleep(100 * time.Millisecond)

		realPath, found = dpr.findPossibleDevice(devicePath)
	}

	return realPath, false, nil
}

func (dpr mappedDevicePathResolver) findPossibleDevice(devicePath string) (string, bool) {
	pathSuffix := strings.Split(devicePath, "/dev/sd")[1]

	possiblePrefixes := []string{
		"/dev/xvd", // Xen
		"/dev/vd",  // KVM
		"/dev/sd",
	}

	for _, prefix := range possiblePrefixes {
		path := prefix + pathSuffix
		if dpr.fs.FileExists(path) {
			return path, true
		}
	}

	return "", false
}
