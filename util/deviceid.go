package util

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/shibukawa/configdir"
	"os"
)

// GetMyID get device mesh Id
func GetMyID() uuid.UUID {
	configDirs := configdir.New("sounddrop", "sounddrop")
	config := configDirs.QueryFolders(configdir.Global)[0]
	filePath := fmt.Sprintf("%s/.uuid", config.Path)

	if id, err := getFromFile(filePath); err == nil {
		return id
	}

	if id, err := create(filePath); err == nil {
		return id
	}

	return uuid.UUID{}
}

func getFromFile(filePath string) (uuid.UUID, error) {
	var err error
	if _, err = os.Stat(filePath); err == nil {
		f, _ := os.Open(filePath)
		b := make([]byte, 36)
		_, err := f.Read(b)
		if err != nil {
			return uuid.UUID{}, err
		}

		return uuid.ParseBytes(b)
	}

	return uuid.UUID{}, err
}

func create(filePath string) (uuid.UUID, error) {
	f, _ := os.Create(filePath)
	UUID, err := uuid.NewRandom()
	if err != nil {
		println("uuid.NewRandom failed")
		return uuid.UUID{}, err
	}
	uuidString := UUID.String()
	println(uuidString)
	_, err = f.WriteString(uuidString)
	if err != nil {
		println("WriteString failed. DeviceID has not been written to disk.")
	}

	return UUID, nil
}
