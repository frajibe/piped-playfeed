package sync

import (
	"errors"
	"fmt"
	"piped-playfeed/config"
	"piped-playfeed/config/model"
	"piped-playfeed/db"
	channelDb "piped-playfeed/db/channel"
	dbCommon "piped-playfeed/db/common"
	videoDb "piped-playfeed/db/video"
	pipedApi "piped-playfeed/piped/api"
	pipedDto "piped-playfeed/piped/dto"
	"piped-playfeed/utils"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var instance *SynchronizationService
var mutex sync.Mutex

type SynchronizationService struct {
}

func GetSynchronizationServiceInstance() *SynchronizationService {
	if instance == nil {
		mutex.Lock()
		defer mutex.Unlock()
		if instance == nil {
			instance = &SynchronizationService{}
		}
	}
	return instance
}

func (syncService *SynchronizationService) Synchronize() error {
	// fetch the user subscriptions
	utils.GetLoggingService().Debug("Fetching subscriptions")
	subProgressBar := utils.CreateInfiniteProgressBar("[1/6] Fetching subscriptions...")
	pipedSubscriptions, err := pipedApi.FetchSubscriptions(config.GetConfigurationServiceInstance().Configuration.Instance, pipedApi.GetToken())
	if err != nil {
		return utils.WrapError("unable to retrieve the subscriptions from the Piped instance", err)
	}
	pipedSubscriptionCount := len(*pipedSubscriptions)
	utils.GetLoggingService().Debug(fmt.Sprintf("%v subscriptions found", pipedSubscriptionCount))
	utils.FinalizeProgressBar(subProgressBar, pipedSubscriptionCount)
	if pipedSubscriptionCount == 0 {
		utils.GetLoggingService().Console("no subscriptions found, stopping the synchronization")
		return nil
	}

	// fetch the subscribed channels
	utils.GetLoggingService().Debug("Fetching playlists")
	channelProgressBar := utils.CreateInfiniteProgressBar("[2/6] Fetching playlists...")
	pipedPlaylists, err := syncService.fetchPlaylists()
	if err != nil {
		return utils.WrapError("unable to retrieve the playlists from the Piped instance", err)
	}
	utils.FinalizeProgressBar(channelProgressBar, len(*pipedPlaylists))

	// sync the db with the existing playlist
	videoRepository := db.GetDatabaseServiceInstance().VideoRepository
	err = syncService.synchronizeDbPlaylists(pipedPlaylists, videoRepository, pipedApi.GetToken())
	if err != nil {
		return utils.WrapError("unable to synchronize the playlists in database", err)
	}

	// fetch the new videos
	channelRepository := db.GetDatabaseServiceInstance().ChannelRepository
	videosProgressBar := utils.CreateInfiniteProgressBar("[4/6] Fetching new videos...")
	newPipedVideos := syncService.gatherSubscriptionsNewVideos(pipedSubscriptions, channelRepository)
	utils.FinalizeProgressBar(videosProgressBar, len(*newPipedVideos))
	if len(*newPipedVideos) == 0 {
		utils.GetLoggingService().Console("no new videos found, stopping the synchronization")
		return nil
	}
	utils.GetLoggingService().Info(fmt.Sprintf("%d new videos found", len(*newPipedVideos)))

	// populate the db with the new videos
	relatedPlaylistNames, err := syncService.indexVideos(newPipedVideos, videoRepository)
	if err != nil {
		return utils.WrapError("unable to index the new videos into the database", err)
	}

	// sync the piped playlists with the db
	err = syncService.syncPipedPlaylists(relatedPlaylistNames, videoRepository)
	if err != nil {
		return utils.WrapError("unable to synchronize the Piped instance playlists", err)
	}
	return nil
}

func (syncService *SynchronizationService) indexVideos(newPipedVideos *[]pipedDto.VideoDto, videoRepository *videoDb.SQLiteVideoRepository) ([]string, error) {
	utils.GetLoggingService().Debug("Indexing new videos")
	var relatedPlaylistNames = make(map[string]struct{})
	playlistPrefix := config.GetConfigurationServiceInstance().Configuration.Synchronization.PlaylistPrefix
	playlistStrategy := config.GetConfigurationServiceInstance().Configuration.Synchronization.Strategy
	progressBar := utils.CreateProgressBar(len(*newPipedVideos), "[5/6] Indexing new videos...")
	for _, newPipedVideo := range *newPipedVideos {
		videoId := pipedApi.ExtractIdFromUrl(newPipedVideo.Url)
		_, err := videoRepository.GetById(videoId)
		// try to add the video into db is not already present
		if err != nil {
			if errors.Is(err, dbCommon.ErrNotExists) {
				// the video is new: create it
				playlistName := syncService.determinePlaylistForVideo(newPipedVideo, playlistPrefix, playlistStrategy)
				_, err = videoRepository.Create(videoDb.SubscriptionVideo{
					Id:       videoId,
					Date:     newPipedVideo.Uploaded,
					Removed:  0,
					Playlist: playlistName,
				})
				if err != nil {
					return nil, utils.WrapError(fmt.Sprintf("Can't create the video in database '%s'", videoId), err)
				}
				relatedPlaylistNames[playlistName] = struct{}{}
			} else {
				return nil, utils.WrapError(fmt.Sprintf("Can't read the video from database '%s'", videoId), err)
			}
		}
		utils.IncrementProgressBar(progressBar)
	}
	utils.FinalizeProgressBar(progressBar, len(*newPipedVideos))

	var uniquePlaylistNames []string
	for key := range relatedPlaylistNames {
		uniquePlaylistNames = append(uniquePlaylistNames, key)
	}
	utils.GetLoggingService().Debug("... indexing done")
	return uniquePlaylistNames, nil
}

func (syncService *SynchronizationService) synchronizeDbPlaylists(pipedPlaylists *[]pipedDto.PlaylistDto, subscriptionVideoRepository *videoDb.SQLiteVideoRepository, userToken string) error {
	// gather the id of the videos that are part of the playlists
	var playlistsVideosIds []string
	progressBar := utils.CreateProgressBar(len(*pipedPlaylists), "[3/6] Indexing playlists...")
	for _, pipedPlaylist := range *pipedPlaylists {
		pipedVideos, err := pipedApi.FetchPlaylistVideos(pipedPlaylist.Id, config.GetConfigurationServiceInstance().Configuration.Instance, userToken)
		if err != nil {
			return utils.WrapError("unable to retrieve the playlists videos", err)
		}
		for _, pipedVideo := range *pipedVideos {
			playlistsVideosIds = append(playlistsVideosIds, pipedApi.ExtractIdFromUrl(pipedVideo.Url))
		}
		utils.IncrementProgressBar(progressBar)
	}
	utils.FinalizeProgressBar(progressBar, len(*pipedPlaylists))

	// tag all the videos that are not part of the playlist as manually removed
	err := subscriptionVideoRepository.SetAllRemoved(&playlistsVideosIds)
	if err != nil {
		utils.GetLoggingService().Warn(utils.WrapError("unable to mark videos as manually removed", err).Error())
	}
	return nil
}

func (syncService *SynchronizationService) gatherSubscriptionsNewVideos(pipedSubscriptions *[]pipedDto.SubscriptionDto, subscriptionChannelRepository *channelDb.SQLiteChannelRepository) *[]pipedDto.VideoDto {
	subscribedPipedVideos := make([]pipedDto.VideoDto, 0, 1000)
	for _, pipedSubscription := range *pipedSubscriptions {
		newPipedVideos, err := syncService.gatherSubscriptionNewVideos(pipedSubscription, subscriptionChannelRepository)
		if err != nil {
			msg := fmt.Sprintf("Unable to retrieve new videos for the channel '%s'", pipedSubscription.Name)
			utils.GetLoggingService().ConsoleWarn(msg)
			utils.GetLoggingService().WarnFromError(utils.WrapError(msg, err))
		}
		subscribedPipedVideos = append(subscribedPipedVideos, *newPipedVideos...)
	}

	// sort them by creation date
	sort.Slice(subscribedPipedVideos, func(v1, v2 int) bool {
		return subscribedPipedVideos[v1].Uploaded > subscribedPipedVideos[v2].Uploaded
	})
	return &subscribedPipedVideos
}

func (syncService *SynchronizationService) gatherSubscriptionNewVideos(pipedSubscription pipedDto.SubscriptionDto, subscriptionChannelRepository *channelDb.SQLiteChannelRepository) (*[]pipedDto.VideoDto, error) {
	utils.GetLoggingService().Debug(fmt.Sprintf("Fetching subscription channel '%s'", pipedSubscription.Name))
	configuration := config.GetConfigurationServiceInstance().Configuration
	pipedChannel, err := pipedApi.FetchChannel(pipedSubscription, configuration.Instance)
	if err != nil {
		return nil, utils.WrapError(fmt.Sprintf("unable to retrieve the channel '%s'", pipedSubscription.Name), err)
	}

	// find the channel in db (create it if needed)
	utils.GetLoggingService().Debug("Looking for channel in database")
	subscriptionChannel, err := subscriptionChannelRepository.GetById(pipedChannel.Id)
	if err != nil {
		if errors.Is(err, dbCommon.ErrNotExists) {
			utils.GetLoggingService().Debug("... channel not found, creating it...")
			subscriptionChannel, err = subscriptionChannelRepository.Create(channelDb.SubscriptionChannel{
				Id:            pipedChannel.Id,
				LastVideoDate: 0,
			})
			if err != nil {
				return nil, utils.WrapError(fmt.Sprintf("unable to create the channel in database: '%s'", pipedSubscription.Name), err)
			}
		} else {
			return nil, utils.WrapError(fmt.Sprintf("unexpected error when fetching the channel from database: '%s'", pipedSubscription.Name), err)
		}
	} else {
		utils.GetLoggingService().Debug("... channel found")
	}

	// determine the oldest date in the past according to the sync conf and the last video seen in the channel
	var syncOldestDateAllowed time.Time
	if strings.EqualFold(configuration.Synchronization.Type, model.SyncDurationType) {
		now := time.Now()
		startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
		switch configuration.Synchronization.Duration.Unit {
		case model.SyncDurationUnitMonth:
			syncOldestDateAllowed = startOfDay.AddDate(0, int(-configuration.Synchronization.Duration.Value), 0)
		case model.SyncDurationUnitDay:
			syncOldestDateAllowed = startOfDay.AddDate(0, 0, int(-configuration.Synchronization.Duration.Value))
		}
	} else {
		// it assumes that the date has already been checked at startup
		syncOldestDateAllowed, _ = time.Parse("2006-01-02", configuration.Synchronization.Date)
	}
	channelOldestDateAllowed := time.UnixMilli(subscriptionChannel.LastVideoDate)
	var oldestDateAllowed time.Time
	if syncOldestDateAllowed.After(channelOldestDateAllowed) {
		oldestDateAllowed = syncOldestDateAllowed
	} else {
		oldestDateAllowed = channelOldestDateAllowed
	}

	utils.GetLoggingService().Debug(fmt.Sprintf("Fetching videos since %s", oldestDateAllowed))
	videos, err := pipedApi.FetchChannelVideos(pipedChannel, oldestDateAllowed, configuration.Instance)
	if err != nil {
		return nil, utils.WrapError(fmt.Sprintf("unable to retrieve the videos for channel '%s'", pipedSubscription.Name), err)
	}
	utils.GetLoggingService().Debug(fmt.Sprintf("... %v found", len(*videos)))

	// update the persisted channel video date
	if len(*videos) != 0 {
		subscriptionChannel.LastVideoDate = (*videos)[0].Uploaded
		if _, err := subscriptionChannelRepository.Update(subscriptionChannel.Id, *subscriptionChannel); err != nil {
			return nil, utils.WrapError(fmt.Sprintf("Unable to update the channel in database: '%s'", pipedSubscription.Name), err)
		}
	}
	return videos, nil
}

func (syncService *SynchronizationService) syncPipedPlaylists(playlistNames []string, subscriptionVideoRepository *videoDb.SQLiteVideoRepository) error {
	// retrieve the playlists to be updated
	utils.GetLoggingService().Debug("Populating playlist...")
	utils.GetLoggingService().ConsoleProgress("[6/6] Populating playlist...")
	playlistsSortedByName, err := syncService.fetchPlaylistsMap()
	if err != nil {
		return err
	}
	for _, playlistName := range playlistNames {
		utils.GetLoggingService().Debug(fmt.Sprintf("%s", playlistName))
		pipedPlaylist, playlistPresent := (*playlistsSortedByName)[playlistName]
		var playlistId string
		if !playlistPresent {
			// create the playlist if missing
			playlist, err := pipedApi.CreatePlaylist(playlistName, config.GetConfigurationServiceInstance().Configuration.Instance, pipedApi.GetToken())
			if err != nil {
				return utils.WrapError("can't create playlist in the piped instance", err)
			}
			playlistId = playlist.PlaylistId
		} else {
			// clear the existing playlist
			err := pipedApi.RemoveAllPlaylistVideos(pipedPlaylist.Id, config.GetConfigurationServiceInstance().Configuration.Instance, pipedApi.GetToken())
			if err != nil {
				return utils.WrapError("can't clear the existing playlist", err)
			}
			playlistId = pipedPlaylist.Id
		}

		// populate the playlist with its videos
		videos, err := subscriptionVideoRepository.GetByPlaylist(playlistName)
		if err != nil {
			return utils.WrapError(fmt.Sprintf("can't read the playlist from database '%s'", playlistName), err)
		}
		progressBar := utils.CreateProgressBar(len(*videos), fmt.Sprintf("'%s'", playlistName))
		for _, video := range *videos {
			utils.IncrementProgressBar(progressBar)
			pipedVideo := pipedDto.VideoDto{
				Url: pipedApi.BuildVideoUrl(video.Id),
			}
			err := pipedApi.AddVideoIntoPlaylist(playlistId, &pipedVideo, config.GetConfigurationServiceInstance().Configuration.Instance, pipedApi.GetToken())
			if err != nil {
				return utils.WrapError(fmt.Sprintf("can't insert video into playlist '%s'", playlistName), err)
			}
		}
	}
	utils.GetLoggingService().Debug("... populating done")
	return nil
}

func (syncService *SynchronizationService) determinePlaylistForVideo(pipedVideo pipedDto.VideoDto, prefix string, playlistCreationStrategy string) string {
	videoDate := time.UnixMilli(pipedVideo.Uploaded)
	var strategySuffix string
	if strings.EqualFold(playlistCreationStrategy, model.PlaylistMonthlyStrategy) {
		strategySuffix = videoDate.Month().String()
	} else {
		_, month := videoDate.ISOWeek()
		strategySuffix = fmt.Sprintf("Week %v", strconv.Itoa(month))
	}
	return fmt.Sprintf("%v%v %v", prefix, videoDate.Year(), strategySuffix)
}

func (syncService *SynchronizationService) fetchPlaylists() (*[]pipedDto.PlaylistDto, error) {
	var filteredPlaylists []pipedDto.PlaylistDto
	prefix := config.GetConfigurationServiceInstance().Configuration.Synchronization.PlaylistPrefix
	pipedPlaylists, err := pipedApi.FetchPlaylists(config.GetConfigurationServiceInstance().Configuration.Instance, pipedApi.GetToken())
	if err != nil {
		return nil, err
	}
	for _, playlist := range *pipedPlaylists {
		if strings.HasPrefix(playlist.Name, prefix) {
			filteredPlaylists = append(filteredPlaylists, playlist)
		}
	}
	return &filteredPlaylists, nil
}

func (syncService *SynchronizationService) fetchPlaylistsMap() (*map[string]pipedDto.PlaylistDto, error) {
	var pipedPlaylistsByName = make(map[string]pipedDto.PlaylistDto)
	pipedPlaylists, err := syncService.fetchPlaylists()
	if err != nil {
		return nil, utils.WrapError("unable to retrieve the playlists", err)
	}
	for _, pipedPlaylist := range *pipedPlaylists {
		pipedPlaylistsByName[pipedPlaylist.Name] = pipedPlaylist
	}
	return &pipedPlaylistsByName, nil
}
