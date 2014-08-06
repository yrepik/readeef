// +build postgres

package readeef

import _ "github.com/lib/pq"

var (
	init_sql_postgres = []string{`
CREATE TABLE IF NOT EXISTS users (
	login TEXT PRIMARY KEY,
	first_name TEXT,
	last_name TEXT,
	email TEXT,
	salt BYTEA,
	hash BYTEA,
	md5_api BYTEA
)`, `
CREATE TABLE IF NOT EXISTS feeds (
	link TEXT PRIMARY KEY,
	title TEXT,
	description TEXT,
	hub_link TEXT,
	update_error TEXT,
	subscribe_error TEXT
)`, `
CREATE TABLE IF NOT EXISTS feed_images (
	id INTEGER PRIMARY KEY,
	feed_link TEXT,
	title TEXT,
	url TEXT,
	width INTEGER,
	height INTEGER,

	FOREIGN KEY(feed_link) REFERENCES feeds(link) ON DELETE CASCADE
)`, `
CREATE TABLE IF NOT EXISTS articles (
	id TEXT,
	feed_link TEXT,
	title TEXT,
	description TEXT,
	link TEXT,
	date TIMESTAMP WITH TIME ZONE,

	PRIMARY KEY(id, feed_link),
	FOREIGN KEY(feed_link) REFERENCES feeds(link) ON DELETE CASCADE
)`, `
CREATE TABLE IF NOT EXISTS users_feeds (
	user_login TEXT,
	feed_link TEXT,

	PRIMARY KEY(user_login, feed_link),
	FOREIGN KEY(user_login) REFERENCES users(login) ON DELETE CASCADE,
	FOREIGN KEY(feed_link) REFERENCES feeds(link) ON DELETE CASCADE
)`, `
CREATE TABLE IF NOT EXISTS users_articles_read (
	user_login TEXT,
	article_id TEXT,
	article_feed_link TEXT,

	PRIMARY KEY(user_login, article_id, article_feed_link),
	FOREIGN KEY(user_login) REFERENCES users(login) ON DELETE CASCADE,
	FOREIGN KEY(article_id, article_feed_link) REFERENCES articles(id, feed_link) ON DELETE CASCADE
)`, `
CREATE TABLE IF NOT EXISTS users_articles_fav (
	user_login TEXT,
	article_id TEXT,
	article_feed_link TEXT,

	PRIMARY KEY(user_login, article_id, article_feed_link),
	FOREIGN KEY(user_login) REFERENCES users(login) ON DELETE CASCADE,
	FOREIGN KEY(article_id, article_feed_link) REFERENCES articles(id, feed_link) ON DELETE CASCADE
)`, `
CREATE TABLE IF NOT EXISTS hubbub_subscriptions (
	link TEXT,
	feed_link TEXT NOT NULL UNIQUE,
	lease_duration INTEGER,
	verification_time TIMESTAMP WITH TIME ZONE,
	subscription_failure BOOLEAN,

	PRIMARY KEY(link),
	FOREIGN KEY(feed_link) REFERENCES feeds(link) ON DELETE CASCADE
)`,
	}
)

func init() {
	init_sql["postgres"] = init_sql_postgres
}