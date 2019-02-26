package util

import (
	"crypto/tls"
	"fmt"
	"github.com/shibukawa/configdir"
	"github.com/syncthing/syncthing/lib/protocol"
	"github.com/syncthing/syncthing/lib/tlsutil"
)

// GetMyID get device mesh Id
func GetMyID() (protocol.DeviceID, error) {
	configDirs := configdir.New("sounddrop", "sounddrop")
	config := configDirs.QueryFolders(configdir.Global)[0]
	configPath := config.Path

	cert, err := tls.LoadX509KeyPair(fmt.Sprintf("%s/%s", configPath, "cert.pem"), fmt.Sprintf("%s/%s", configPath, "key.pem"))
	if err != nil {
		err := config.CreateParentDir(fmt.Sprintf("%s/%s", configPath, "cert.pem"))
		if err != nil {
			return protocol.DeviceID{}, err
		}
		cert, err = tlsutil.NewCertificate(fmt.Sprintf("%s/%s", configPath, "cert.pem"), fmt.Sprintf("%s/%s", configPath, "key.pem"), "sounddrop")
		if err != nil {
			return protocol.DeviceID{}, err
		}
	}

	return protocol.NewDeviceID(cert.Certificate[0]), nil
}
