package sqlite3

import (
	"fmt"
	"time"

	"github.com/urandom/readeef/content/sql/db"
	"github.com/urandom/readeef/content/sql/db/base"
)

type Helper struct {
	base.Helper
}

func (h Helper) InitSQL() []string {
	return initSQL
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

	helper.Set("create_feed_article", createFeedArticle)
	helper.Set("get_user_feeds", getUserFeeds)
	helper.Set("get_user_tag_feeds", getUserTagFeeds)
	helper.Set("get_latest_feed_articles", getLatestFeedArticles)

	db.Register("sqlite3", helper)
}

const (
	// Casting to timestamp produces only the year
	createFeedArticle = `
INSERT INTO articles(feed_id, link, guid, title, description, date)
	SELECT $1, $2, $3, $4, $5, $6 EXCEPT
		SELECT feed_id, link, $3, $4, $5, $6
		FROM articles WHERE feed_id = $1 AND link = $2
`
	getUserFeeds = `
SELECT f.id, f.link, f.title, f.description, f.link, f.hub_link, f.site_link, f.update_error, f.subscribe_error
FROM feeds f, users_feeds uf
WHERE f.id = uf.feed_id
	AND uf.user_login = $1
ORDER BY f.title COLLATE NOCASE
`
	getUserTagFeeds = `
SELECT f.id, f.link, f.title, f.description, f.link, f.hub_link, f.site_link, f.update_error, f.subscribe_error
FROM feeds f, users_feeds_tags uft
WHERE f.id = uft.feed_id
	AND uft.user_login = $1 AND uft.tag = $2
ORDER BY f.title COLLATE NOCASE
`

	getLatestFeedArticles = `
SELECT a.feed_id, a.id, a.title, a.description, a.link, a.date, a.guid
FROM articles a
WHERE a.feed_id = $1 AND a.date > DATE('NOW', '-5 days')
`

	upgrade1To2MergeReadAndFav = `
INSERT INTO users_articles_states
SELECT ar.user_login, ar.article_id, 1 as read,
    CASE WHEN af.article_id IS NULL THEN 0 ELSE 1 END AS favorite
FROM users_articles_read ar LEFT OUTER JOIN users_articles_fav af
    ON ar.article_id = af.article_id AND ar.user_login = af.user_login
UNION ALL
SELECT af.user_login, af.article_id,
    CASE WHEN ar.article_id IS NULL THEN 0 ELSE 1 END AS read, 1 as favorite
FROM users_articles_fav af LEFT OUTER JOIN users_articles_read ar
    ON af.user_login = ar.user_login AND af.article_id = ar.article_id
WHERE ar.article_id IS NULL
ORDER BY ar.article_id, af.article_id
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
