package sql

import (
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/base"
	"github.com/urandom/readeef/content/info"
	"github.com/urandom/readeef/db"
	"github.com/urandom/webfw"
)

type Repo struct {
	base.Repo
	logger webfw.Logger

	db *db.DB
}

func NewRepo(db *db.DB, logger webfw.Logger) *Repo {
	return &Repo{db: db, logger: logger}
}

func (r *Repo) UserByLogin(login info.Login) (u content.User) {
	if r.Err() != nil {
		return
	}

	r.logger.Infof("Getting user '%s'\n", login)

	var info info.User
	if err := r.db.Get(&info, db.SQL("get_user"), login); err != nil {
		r.SetErr(err)
		return
	}

	info.Login = login
	u.Set(info)

	if u.Err() != nil {
		r.SetErr(u.Err())
	}

	return
}

func (r *Repo) UserByMD5Api(md5 []byte) (u content.User) {
	if r.Err() != nil {
		return
	}

	r.logger.Infof("Getting user using md5 api field '%v'\n", md5)

	var info info.User
	if err := r.db.Get(&info, db.SQL("get_user_by_md5_api"), md5); err != nil {
		r.SetErr(err)
		return
	}

	info.MD5API = md5
	u.Set(info)

	if u.Err() != nil {
		r.SetErr(u.Err())
	}

	return
}

func (r *Repo) AllUsers() (users []content.User) {
	if r.Err() != nil {
		return
	}

	r.logger.Infoln("Getting all users")

	var info []info.User
	if err := r.db.Select(&info, db.SQL("get_users")); err != nil {
		r.SetErr(err)
		return
	}

	users = make([]content.User, len(info))

	for i := range info {
		users[i].Set(info[i])
		if users[i].Err() != nil {
			r.SetErr(users[i].Err())
			return
		}
	}

	return
}

func (r *Repo) FeedById(id info.FeedId) (f content.Feed) {
	if r.Err() != nil {
		return
	}

	r.logger.Infof("Getting feed '%d'\n", id)

	i := info.Feed{}
	if err := r.db.Get(&i, db.SQL("get_feed"), id); err != nil {
		r.SetErr(err)
		return
	}

	i.Id = id
	f.Set(i)

	return
}

func (r *Repo) FeedByLink(link string) (f content.Feed) {
	if r.Err() != nil {
		return
	}

	r.logger.Infof("Getting feed by link '%s'\n", link)

	i := info.Feed{}
	if err := r.db.Get(&i, db.SQL("get_feed_by_link"), link); err != nil {
		r.SetErr(err)
		return
	}

	i.Link = link
	f.Set(i)

	return
}

func (r *Repo) AllFeeds() (feeds []content.Feed) {
	if r.Err() != nil {
		return
	}

	r.logger.Infoln("Getting all feeds")

	var info []info.Feed
	if err := r.db.Select(&info, db.SQL("get_feeds")); err != nil {
		r.SetErr(err)
		return
	}

	feeds = make([]content.Feed, len(info))

	for i := range info {
		feeds[i].Set(info[i])
	}

	return
}

func (r *Repo) AllUnsubscribedFeeds() (feeds []content.Feed) {
	if r.Err() != nil {
		return
	}

	r.logger.Infoln("Getting all unsubscribed feeds")

	var info []info.Feed
	if err := r.db.Select(&info, db.SQL("get_unsubscribed_feeds")); err != nil {
		r.SetErr(err)
		return
	}

	feeds = make([]content.Feed, len(info))

	for i := range info {
		feeds[i].Set(info[i])
	}

	return
}

func (r *Repo) AllSubscriptions() (s []content.Subscription) {
	if r.Err() != nil {
		return
	}

	r.logger.Infoln("Getting all subscriptions")

	var info []info.Subscription
	if err := r.db.Select(&info, db.SQL("get_hubbub_subscriptions")); err != nil {
		r.SetErr(err)
		return
	}

	s = make([]content.Subscription, len(info))

	for i := range info {
		s[i].Set(info[i])
	}

	return
}

func init() {
	db.SetSQL("get_user", getUser)
	db.SetSQL("get_user_by_md5_api", getUser)
	db.SetSQL("get_users", getUsers)
	db.SetSQL("get_feed", getFeed)
	db.SetSQL("get_feed_by_link", getFeedByLink)
	db.SetSQL("get_feeds", getFeeds)
	db.SetSQL("get_unsubscribed_feeds", getUnsubscribedFeeds)
	db.SetSQL("get_hubbub_subscriptions", getHubbubSubscriptions)
}

func pagingLimit(paging []int) (int, int) {
	limit := 50
	offset := 0

	if len(paging) > 0 {
		limit = paging[0]
		if len(paging) > 1 {
			offset = paging[1]
		}
	}

	return limit, offset
}

const (
	getUser         = `SELECT first_name, last_name, email, admin, active, profile_data, hash_type, salt, hash, md5_api FROM users WHERE login = $1`
	getUserByMD5Api = `SELECT login, first_name, last_name, email, admin, active, profile_data, hash_type, salt, hash FROM users WHERE md5_api = $1`
	getUsers        = `SELECT login, first_name, last_name, email, admin, active, profile_data, hash_type, salt, hash, md5_api FROM users`

	getFeed              = `SELECT link, title, description, hub_link, site_link, update_error, subscribe_error FROM feeds WHERE id = $1`
	getFeedByLink        = `SELECT id, title, description, hub_link, site_link, update_error, subscribe_error FROM feeds WHERE link = $1`
	getFeeds             = `SELECT id, link, title, description, hub_link, site_link, update_error, subscribe_error FROM feeds`
	getUnsubscribedFeeds = `
SELECT f.id, f.link, f.title, f.description, f.hub_link, f.site_link, f.update_error, f.subscribe_error
	FROM feeds f LEFT OUTER JOIN hubbub_subscriptions hs
	ON f.id = hs.feed_id AND hs.subscription_failure = '1'
	ORDER BY f.title
`
	getHubbubSubscriptions = `
SELECT link, feed_id, lease_duration, verification_time, subscription_failure
	FROM hubbub_subscriptions`
)
