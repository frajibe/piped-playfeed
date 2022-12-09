package db

import (
	"database/sql"
	"fmt"
	"piped-playfeed/config"
	channelDb "piped-playfeed/db/channel"
	videoDb "piped-playfeed/db/video"
	"piped-playfeed/utils"
	"sync"
)

var instance *DatabaseService
var mutex sync.Mutex

type DatabaseService struct {
	ChannelRepository *channelDb.SQLiteChannelRepository
	VideoRepository   *videoDb.SQLiteVideoRepository
}

func GetDatabaseServiceInstance() *DatabaseService {
	if instance == nil {
		mutex.Lock()
		defer mutex.Unlock()
		if instance == nil {
			instance = &DatabaseService{}
		}
	}
	return instance
}

func (dbService *DatabaseService) Init() error {
	// create a connection to the db
	configuration := config.GetConfigurationServiceInstance().Configuration
	db, err := sql.Open("sqlite3", configuration.Database)
	if err != nil {
		return utils.WrapError(fmt.Sprintf("can't open the database: '%s'", configuration.Database), err)
	}

	dbService.ChannelRepository = channelDb.NewSQLiteRepository(db)
	if err := dbService.ChannelRepository.Migrate(); err != nil {
		return utils.WrapError("Unable to init the 'channel' table", err)
	}
	dbService.VideoRepository = videoDb.NewSQLiteRepository(db)
	if err := dbService.VideoRepository.Migrate(); err != nil {
		return utils.WrapError("Unable to init the 'video' table", err)
	}
	return nil
}
