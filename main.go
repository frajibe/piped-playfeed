package main

import (
	"flag"
	"fmt"
	"github.com/frajibe/piped-playfeed/config"
	"github.com/frajibe/piped-playfeed/db"
	"github.com/frajibe/piped-playfeed/lock"
	pipedApi "github.com/frajibe/piped-playfeed/piped/api"
	"github.com/frajibe/piped-playfeed/settings"
	"github.com/frajibe/piped-playfeed/sync"
	"github.com/frajibe/piped-playfeed/utils"
	"os"
)

var helpFlag = flag.Bool("help", false, "Show help")
var configFlag = flag.String("conf", "piped-playfeed-conf.json", "Provide the path to the configuration file")
var debugFlag = flag.Bool("debug", false, "Enable debug logging")
var logFlag = flag.String("log", "piped-playfeed-log.json", "Provide the path to the output log file")
var silentFlag = flag.Bool("silent", false, "Hide progress in console")
var syncFlag = flag.Bool("sync", false, "Action: synchronize the playlists accordingly to the subscriptions")
var versionFlag = flag.Bool("version", false, "Show version")

func main() {
	// input args
	parseArguments()

	// lock file
	err := lock.CreateLockFile()
	if err != nil {
		utils.GetLoggingService().FatalFromError(utils.WrapError("unable to create the lock file", err))
	}

	// read the configuration
	confService := config.GetConfigurationServiceInstance()
	configuration, err := confService.Init(*configFlag)
	if err != nil {
		utils.GetLoggingService().FatalFromError(utils.WrapError("unable to get the configuration", err))
	}

	// init the DB if needed
	databaseService := db.GetDatabaseServiceInstance()
	err = databaseService.Init()
	if err != nil {
		utils.GetLoggingService().FatalFromError(utils.WrapError("unable to use a local database", err))
	}

	// login
	err = pipedApi.Login(configuration.Account.Username, configuration.Account.Password, configuration.Instance)
	if err != nil {
		utils.GetLoggingService().FatalFromError(utils.WrapError("unable to authenticate on the Piped instance", err))
	}

	// launch the synchronization if requested
	if settings.GetSettingsService().SynchronizationRequested {
		syncService := sync.GetSynchronizationServiceInstance()
		err = syncService.Synchronize()
		if err != nil {
			utils.GetLoggingService().FatalFromError(utils.WrapError("failed to synchronize", err))
		}
	}

	// ends up the app
	finalize()
}

func parseArguments() {
	// parse the args and let Flag decides if the args are provided
	flag.Parse()

	// init the logging service
	utils.GetLoggingService().InitializeLogger(*logFlag, *debugFlag, finalize)

	// is help needed?
	if *helpFlag {
		flag.Usage()
		os.Exit(0)
	}

	// is version requested?
	if *versionFlag {
		utils.GetLoggingService().Console("v1.1.1")
		os.Exit(0)
	}

	// apply the silent mode if requested
	settings.GetSettingsService().SilentMode = *silentFlag

	// check the conf file path.
	configFile := *configFlag
	if _, err := os.Stat(configFile); err != nil {
		utils.GetLoggingService().FatalFromError(
			utils.WrapError(
				fmt.Sprintf("The configuration file doesn't exist: '%s'", configFile),
				err),
		)
	}

	// ensure that the sync action is requested
	if !*syncFlag {
		flag.Usage()
		os.Exit(0)
	}
	settings.GetSettingsService().SynchronizationRequested = true
}

func finalize() {
	err := utils.GetLoggingService().SyncLogger()
	if err != nil {
		utils.GetLoggingService().ConsoleWarn(fmt.Errorf("logger badly flushed, the log file may be incomplete\n%w", err).Error())
	}
	err = lock.DeleteLockFile()
	if err != nil {
		utils.GetLoggingService().ConsoleFatal(fmt.Errorf("unable to delete the lock file: %w", err).Error())
	}
}
