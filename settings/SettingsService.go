package settings

import (
	"sync"
)

var instance *SettingsService
var mutex sync.Mutex

type SettingsService struct {
	SilentMode               bool
	SynchronizationRequested bool
}

func GetSettingsService() *SettingsService {
	if instance == nil {
		mutex.Lock()
		defer mutex.Unlock()
		if instance == nil {
			instance = &SettingsService{}
		}
	}
	return instance
}
