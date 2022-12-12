package video

import (
	"database/sql"
	"errors"
	"fmt"
	dbCommon "piped-playfeed/db/common"
	"strings"

	"github.com/mattn/go-sqlite3"
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
        date INTEGER,
        removed INTEGER,
        playlist TEXT
    );
    `

	_, err := r.db.Exec(query)
	return err
}

func (r *SQLiteVideoRepository) Create(subscriptionVideo SubscriptionVideo) (*SubscriptionVideo, error) {
	_, err := r.db.Exec("INSERT INTO subscriptions_videos(id, date, removed, playlist) values(?, ?, ?, ?)", subscriptionVideo.Id, subscriptionVideo.Date, subscriptionVideo.Removed, subscriptionVideo.Playlist)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
				return nil, dbCommon.ErrDuplicate
			}
		}
		return nil, err
	}
	return &subscriptionVideo, nil
}

func (r *SQLiteVideoRepository) GetById(id string) (*SubscriptionVideo, error) {
	row := r.db.QueryRow("SELECT * FROM subscriptions_videos WHERE id = ?", id)

	var subscriptionVideo SubscriptionVideo
	if err := row.Scan(&subscriptionVideo.Id, &subscriptionVideo.Date, &subscriptionVideo.Removed, &subscriptionVideo.Playlist); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, dbCommon.ErrNotExists
		}
		return nil, err
	}
	return &subscriptionVideo, nil
}

func (r *SQLiteVideoRepository) GetByPlaylist(playlistName string) (*[]SubscriptionVideo, error) {
	rows, err := r.db.Query("SELECT * FROM subscriptions_videos WHERE playlist = ? AND removed = 0 ORDER BY date DESC", playlistName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []SubscriptionVideo
	for rows.Next() {
		var video SubscriptionVideo
		if err := rows.Scan(&video.Id, &video.Date, &video.Removed, &video.Playlist); err != nil {
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
	res, err := r.db.Exec("UPDATE subscriptions_videos SET date = ?, removed = ?, playlist = ? WHERE id = ?", updated.Date, updated.Removed, updated.Playlist, updated.Id)
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

func (r *SQLiteVideoRepository) SetAllRemoved(excludedIds *[]string) error {
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
