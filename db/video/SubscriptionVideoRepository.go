package video

import (
	"database/sql"
	"errors"
	"fmt"
	dbCommon "github.com/frajibe/piped-playfeed/db/common"
	"strings"
)

type SQLiteVideoRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(db *sql.DB) *SQLiteVideoRepository {
	return &SQLiteVideoRepository{
		db: db,
	}
}

func (r *SQLiteVideoRepository) Migrate() error {
	query := `
    CREATE TABLE IF NOT EXISTS subscriptions_videos(
        id TEXT PRIMARY KEY,
        uploadDate TEXT,
        uploaded INTEGER,
        removed INTEGER,
        playlist TEXT
    );
    `

	_, err := r.db.Exec(query)
	return err
}

func (r *SQLiteVideoRepository) Create(subscriptionVideo SubscriptionVideo) (*SubscriptionVideo, error) {
	_, err := r.db.Exec("INSERT INTO subscriptions_videos(id, uploadDate, uploaded, removed, playlist) values(?, ?, ?, ?, ?)", subscriptionVideo.Id, subscriptionVideo.UploadDate, subscriptionVideo.Uploaded, subscriptionVideo.Removed, subscriptionVideo.Playlist)
	if err != nil {
		return nil, err
	}
	return &subscriptionVideo, nil
}

func (r *SQLiteVideoRepository) Exists(id string) (bool, error) {
	if err := r.db.QueryRow("SELECT id FROM subscriptions_videos WHERE id = ?", id).Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *SQLiteVideoRepository) GetById(id string) (*SubscriptionVideo, error) {
	row := r.db.QueryRow("SELECT * FROM subscriptions_videos WHERE id = ?", id)

	var subscriptionVideo SubscriptionVideo
	if err := row.Scan(&subscriptionVideo.Id, &subscriptionVideo.UploadDate, &subscriptionVideo.Uploaded, &subscriptionVideo.Removed, &subscriptionVideo.Playlist); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, dbCommon.ErrNotExists
		}
		return nil, err
	}
	return &subscriptionVideo, nil
}

func (r *SQLiteVideoRepository) GetByPlaylist(playlistName string) (*[]SubscriptionVideo, error) {
	rows, err := r.db.Query("SELECT *, unixepoch(uploadDate)*1000 as max_date FROM subscriptions_videos WHERE playlist = ? AND removed = 0 ORDER BY max(uploaded, max_date) DESC", playlistName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []SubscriptionVideo
	var maxDate int64
	for rows.Next() {
		var video SubscriptionVideo
		if err := rows.Scan(&video.Id, &video.UploadDate, &video.Uploaded, &video.Removed, &video.Playlist, &maxDate); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, dbCommon.ErrNotExists
			}
			return nil, err
		}
		videos = append(videos, video)
	}

	return &videos, nil
}

func (r *SQLiteVideoRepository) Update(id string, updated SubscriptionVideo) (*SubscriptionVideo, error) {
	if len(id) == 0 {
		return nil, errors.New("invalid updated ID")
	}
	res, err := r.db.Exec("UPDATE subscriptions_videos SET uploadDate = ?, uploaded = ?, removed = ?, playlist = ? WHERE id = ?", updated.UploadDate, updated.Uploaded, updated.Removed, updated.Playlist, updated.Id)
	if err != nil {
		return nil, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	if rowsAffected == 0 {
		return nil, dbCommon.ErrUpdateFailed
	}

	return &updated, nil
}

func (r *SQLiteVideoRepository) SetAllRemovedExcept(excludedIds *[]string) error {
	var excludedIdsStr string
	if len(*excludedIds) != 0 {
		excludedIdsStr = fmt.Sprintf("('%v')", strings.Join(*excludedIds, "', '"))
	} else {
		excludedIdsStr = "()"
	}

	_, err := r.db.Exec("UPDATE subscriptions_videos SET removed = ? WHERE id NOT IN "+excludedIdsStr, 1)
	if err != nil {
		return err
	}
	return nil
}
