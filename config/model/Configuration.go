package model

import "strings"

var defaultDatabaseName = "piped-playfeed.db"

type Configuration struct {
	Instance        string  `validate:"required"`
	Account         Account `validate:"required"`
	Database        string
	Synchronization Synchronization `validate:"-"`
}

func (configuration *Configuration) SetDefaults() {
	if strings.TrimSpace(configuration.Database) == "" {
		configuration.Database = defaultDatabaseName
	}
	configuration.Synchronization.SetDefaults()
}
