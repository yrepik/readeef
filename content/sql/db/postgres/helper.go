package postgres

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/urandom/readeef/content/sql/db"
	"github.com/urandom/readeef/content/sql/db/base"
)

type Helper struct {
	base.Helper
}

func (h Helper) InitSQL() []string {
	return initSQL
}

func (h Helper) CreateWithId(tx *sqlx.Tx, name string, args ...interface{}) (int64, error) {
	var id int64

	sql := h.SQL(name)
	if sql == "" {
		panic("No statement registered under " + name)
	}
	sql += " RETURNING id"

	stmt, err := tx.Preparex(sql)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	err = stmt.QueryRow(args...).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (h Helper) Upgrade(db *db.DB, old, new int) error {
	for old < new {
		var err error
		switch old {
		case 1:
			err = upgrade1to2(db)
		case 2:
			err = upgrade2to3(db)
		}

		if err != nil {
			return fmt.Errorf("Error upgrading db from %d to %d: %v\n", old, new, err)
		}
		old++
	}

	return nil
}

func upgrade1to2(db *db.DB) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(upgrade1To2MergeReadAndFav)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DROP TABLE users_articles_read")
	if err != nil {
		return err
	}

	_, err = tx.Exec("DROP TABLE users_articles_fav")
	if err != nil {
		return err
	}

	return tx.Commit()
}

func upgrade2to3(db *db.DB) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	t := time.Now().AddDate(0, -1, 0)
	_, err = tx.Exec(upgrade2To3SplitStatesToUnread, t)
	if err != nil {
		return err
	}

	_, err = tx.Exec(upgrade2To3SplitStatesToFavorite)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DROP TABLE users_articles_states")
	if err != nil {
		return err
	}

	return tx.Commit()
}

func init() {
	helper := &Helper{Helper: base.NewHelper()}

	helper.Set("get_user_feeds", getUserFeeds)
	helper.Set("get_user_tag_feeds", getUserTagFeeds)

	db.Register("postgres", helper)
}

const (
	getUserFeeds = `
SELECT f.id, f.link, f.title, f.description, f.link, f.hub_link, f.site_link, f.update_error, f.subscribe_error
FROM feeds f, users_feeds uf
WHERE f.id = uf.feed_id
	AND uf.user_login = $1
ORDER BY f.title COLLATE "default"
`
	getUserTagFeeds = `
SELECT f.id, f.link, f.title, f.description, f.link, f.hub_link, f.site_link, f.update_error, f.subscribe_error
FROM feeds f, users_feeds_tags uft
WHERE f.id = uft.feed_id
	AND uft.user_login = $1 AND uft.tag = $2
ORDER BY f.title COLLATE "default"
`

	upgrade1To2MergeReadAndFav = `
INSERT INTO users_articles_states
SELECT COALESCE(ar.user_login, af.user_login), COALESCE(ar.article_id, af.article_id),
	CASE WHEN ar.article_id IS NULL THEN 'f'::BOOLEAN ELSE 't'::BOOLEAN END AS read,
	CASE WHEN af.article_id IS NULL THEN 'f'::BOOLEAN ELSE 't'::BOOLEAN END AS favorite
FROM users_articles_read ar FULL OUTER JOIN users_articles_fav af
	ON ar.article_id = af.article_id AND ar.user_login = af.user_login
ORDER BY ar.article_id
`

	upgrade2To3SplitStatesToUnread = `
INSERT INTO users_articles_unread (user_login, article_id, insert_date)
SELECT uas.user_login, uas.article_id, a.date
FROM users_articles_states uas INNER JOIN articles a
	ON uas.article_id = a.id
WHERE NOT uas.read AND a.date > $1
`

	upgrade2To3SplitStatesToFavorite = `
INSERT INTO users_articles_favorite (user_login, article_id)
SELECT uas.user_login, uas.article_id
FROM users_articles_states uas
WHERE uas.favorite
`
)
