package channel

import (
	"database/sql"
	"errors"
)

type SQLiteChannelRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(db *sql.DB) *SQLiteChannelRepository {
	return &SQLiteChannelRepository{
		db: db,
	}
}

func (r *SQLiteChannelRepository) Migrate() error {
	query := `
    CREATE TABLE IF NOT EXISTS subscriptions_channels(
        id TEXT PRIMARY KEY,
        lastVideoDate INTEGER
    );
    `

	_, err := r.db.Exec(query)
	return err
}

func (r *SQLiteChannelRepository) Create(subscriptionChannel SubscriptionChannel) (*SubscriptionChannel, error) {
	_, err := r.db.Exec("INSERT INTO subscriptions_channels(id, lastVideoDate) values(?, ?)", subscriptionChannel.Id, subscriptionChannel.LastVideoDate)
	if err != nil {
		return nil, err
	}
	return &subscriptionChannel, nil
}

func (r *SQLiteChannelRepository) GetById(id string) (*SubscriptionChannel, error) {
	row := r.db.QueryRow("SELECT * FROM subscriptions_channels WHERE id = ?", id)

	var subscriptionChannel SubscriptionChannel
	if err := row.Scan(&subscriptionChannel.Id, &subscriptionChannel.LastVideoDate); err != nil {
		return nil, err
	}
	return &subscriptionChannel, nil
}

func (r *SQLiteChannelRepository) Update(id string, updated SubscriptionChannel) (*SubscriptionChannel, error) {
	if len(id) == 0 {
		return nil, errors.New("invalid updated ID")
	}
	res, err := r.db.Exec("UPDATE subscriptions_channels SET lastVideoDate = ? WHERE id = ?", updated.LastVideoDate, updated.Id)
	if err != nil {
		return nil, err
	}

	_, err = res.RowsAffected()
	if err != nil {
		return nil, err
	}

	return &updated, nil
}

//func (r *SQLiteRepository) Delete(id string) error {
//	res, err := r.db.Exec("DELETE FROM subscriptions_channels WHERE id = ?", id)
//	if err != nil {
//		return err
//	}
//
//	rowsAffected, err := res.RowsAffected()
//	if err != nil {
//		return err
//	}
//
//	if rowsAffected == 0 {
//		return ErrDeleteFailed
//	}
//
//	return err
//}
