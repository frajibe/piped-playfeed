package config

import (
	"encoding/json"
	"fmt"
	"github.com/frajibe/piped-playfeed/config/model"
	"github.com/frajibe/piped-playfeed/settings"
	"github.com/frajibe/piped-playfeed/utils"
	"github.com/go-playground/validator/v10"
	"os"
	"strings"
	"sync"
	"time"
)

var instance *ConfigurationService
var mutex sync.Mutex

type ConfigurationService struct {
	Configuration model.Configuration
}

func GetConfigurationServiceInstance() *ConfigurationService {
	if instance == nil {
		mutex.Lock()
		defer mutex.Unlock()
		if instance == nil {
			instance = &ConfigurationService{}
		}
	}
	return instance
}

func (confService *ConfigurationService) Init(filePath string) (*model.Configuration, error) {
	// unmarshall json
	err := confService.parseFile(filePath)
	if err != nil {
		return nil, err
	}

	// set default values if needed
	confService.Configuration.SetDefaults()

	// ensure that the content is valid
	err = confService.checkContent()
	if err != nil {
		return nil, utils.WrapError("validation failed", err)
	}
	return &confService.Configuration, nil
}

func (confService *ConfigurationService) parseFile(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return utils.WrapError(fmt.Sprintf("can't open the configuration file: '%s'", filePath), err)
	}
	err = json.Unmarshal(content, &confService.Configuration)
	if err != nil {
		return utils.WrapError(fmt.Sprintf("can't parse the configuration file: '%s'", filePath), err)
	}
	return nil
}

func (confService *ConfigurationService) checkContent() error {
	validate := validator.New()
	err := validate.Struct(confService.Configuration)
	if err != nil {
		return err
	}
	if settings.GetSettingsService().SynchronizationRequested {
		// workaround for https://github.com/go-playground/validator/issues/908 since there is no "skip_unless"
		// the synchronization struct is reduced according to the sync type
		var synchronizationSubset = model.Synchronization{
			Strategy:       confService.Configuration.Synchronization.Strategy,
			PlaylistPrefix: confService.Configuration.Synchronization.PlaylistPrefix,
			Type:           confService.Configuration.Synchronization.Type,
		}
		if strings.EqualFold(synchronizationSubset.Type, model.SyncDurationType) {
			synchronizationSubset.Duration = confService.Configuration.Synchronization.Duration
			synchronizationSubset.Date = "2006-01-02"
		} else if strings.EqualFold(synchronizationSubset.Type, model.SyncDateType) {
			synchronizationSubset.Date = confService.Configuration.Synchronization.Date
			synchronizationSubset.Duration = model.Duration{
				Unit:  "month",
				Value: 1,
			}
		} else {
			synchronizationSubset.Duration = confService.Configuration.Synchronization.Duration
			synchronizationSubset.Date = confService.Configuration.Synchronization.Date
		}

		validate.RegisterValidation("dateinpast", pastDateValidation)
		err = validate.Struct(synchronizationSubset)
		if err != nil {
			//for _, err := range err.(validator.ValidationErrors) {
			//	fmt.Println(err.Namespace())
			//	fmt.Println(err.Field())
			//	fmt.Println(err.StructNamespace())
			//	fmt.Println(err.StructField())
			//	fmt.Println(err.Tag())
			//	fmt.Println(err.ActualTag())
			//	fmt.Println(err.Kind())
			//	fmt.Println(err.Type())
			//	fmt.Println(err.Value())
			//	fmt.Println(err.Param())
			//	fmt.Println()
			//}
			return err
		}
	}
	return nil
}

func pastDateValidation(fl validator.FieldLevel) bool {
	date, err := time.Parse("2006-01-02", fl.Field().String())
	// the date format is checked from another built-in validator, here we just check its content
	if err == nil && !date.Before(time.Now().Local()) {
		return false
	}
	return true
}
