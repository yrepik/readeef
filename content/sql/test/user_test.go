package test

import (
	"testing"
	"time"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/base"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/readeef/tests"
)

func TestUser(t *testing.T) {
	u := repo.User()

	tests.CheckBool(t, false, u.HasErr(), u.Err())

	u.Update()
	tests.CheckBool(t, true, u.HasErr())

	err := u.Err()
	_, ok := err.(base.ValidationError)
	tests.CheckBool(t, true, ok, err)

	u.Data(data.User{Login: data.Login("login")})

	tests.CheckBool(t, false, u.HasErr(), u.Err())

	u.Update()
	tests.CheckBool(t, false, u.HasErr(), u.Err())

	u2 := repo.UserByLogin(data.Login("login"))
	tests.CheckBool(t, false, u2.HasErr(), u2.Err())
	tests.CheckString(t, "login", string(u2.Data().Login))

	u.Delete()
	tests.CheckBool(t, false, u.HasErr(), u.Err())

	u2 = repo.UserByLogin(data.Login("login"))
	tests.CheckBool(t, true, u2.HasErr())
	tests.CheckBool(t, true, u2.Err() == content.ErrNoContent)

	u = createUser(data.User{Login: data.Login("login")})

	now := time.Now()
	uf := createUserFeed(u, data.Feed{Link: "http://sugr.org/en/sitemap.xml", Title: "User feed 1"})
	uf.AddArticles([]content.Article{
		createArticle(data.Article{Id: 11, Title: "article1", Date: now, Link: "http://sugr.org/bg/products/gearshift"}),
		createArticle(data.Article{Id: 12, Title: "article2", Date: now.Add(2 * time.Hour), Link: "http://sugr.org/bg/products/readeef"}),
		createArticle(data.Article{Id: 13, Title: "article3", Date: now.Add(-3 * time.Hour), Link: "http://sugr.org/bg/about/us"}),
	})

	u.AddFeed(uf)

	tests.CheckBool(t, false, uf.HasErr(), uf.Err())
	tests.CheckInt64(t, 1, int64(len(u.AllFeeds())))
	tests.CheckString(t, "http://sugr.org", u.AllFeeds()[0].Data().Link)
	tests.CheckString(t, "User feed 1", u.AllFeeds()[0].Data().Title)

	a := u.ArticleById(10)
	tests.CheckBool(t, true, a.Err() == content.ErrNoContent)

	a = u.ArticleById(11)
	tests.CheckBool(t, false, a.HasErr(), a.Err())

	tests.CheckString(t, "article1", a.Data().Title)

	a2 := u.ArticlesById([]data.ArticleId{10, 11, 12})
	tests.CheckBool(t, false, u.HasErr(), u.Err())

	tests.CheckInt64(t, 2, int64(len(a2)))

	for i := range a2 {
		d := a2[i].Data()
		switch d.Title {
		case "article1":
		case "article2":
		default:
			tests.CheckBool(t, false, true, "Unknown article")
		}
	}

	ua := u.Articles()
	tests.CheckBool(t, false, u.HasErr(), u.Err())

	tests.CheckInt64(t, 3, int64(len(ua)))

	tests.CheckInt64(t, 1, int64(ua[0].Data().Id))
	tests.CheckString(t, "article2", ua[1].Data().Title)
	tests.CheckBool(t, true, now.Add(-3*time.Hour).Equal(ua[2].Data().Date))

	u.SortingByDate()
	ua = u.Articles()

	tests.CheckInt64(t, 3, int64(ua[0].Data().Id))
	tests.CheckString(t, "article1", ua[1].Data().Title)
	tests.CheckBool(t, true, now.Add(2*time.Hour).Equal(ua[2].Data().Date))

	u.Reverse()
	ua = u.Articles()

	tests.CheckInt64(t, 2, int64(ua[0].Data().Id))
	tests.CheckString(t, "article1", ua[1].Data().Title)
	tests.CheckBool(t, true, now.Add(-3*time.Hour).Equal(ua[2].Data().Date))

	ua[0].Read(true)

	u.Reverse()
	u.DefaultSorting()

	ua = u.UnreadArticles()
	tests.CheckBool(t, false, u.HasErr(), u.Err())
	tests.CheckInt64(t, 2, int64(len(ua)))

	tests.CheckInt64(t, 1, int64(ua[0].Data().Id))
	tests.CheckString(t, "article3", ua[1].Data().Title)

	u.ArticleById(data.ArticleId(2)).Read(false)

	ua = u.UnreadArticles()
	tests.CheckInt64(t, 3, int64(len(ua)))

	u.ReadBefore(now.Add(time.Minute), true)
	tests.CheckBool(t, false, u.HasErr(), u.Err())

	ua = u.UnreadArticles()
	tests.CheckBool(t, false, u.HasErr(), u.Err())
	tests.CheckInt64(t, 1, int64(len(ua)))
	tests.CheckInt64(t, 2, int64(ua[0].Data().Id))

	asc1 := createArticleScores(data.ArticleScores{ArticleId: 1, Score1: 2, Score2: 2})
	asc2 := createArticleScores(data.ArticleScores{ArticleId: 2, Score1: 1, Score2: 3})

	sa := u.ScoredArticles(now.Add(-20*time.Hour), now.Add(20*time.Hour))

	tests.CheckBool(t, false, u.HasErr(), u.Err())
	tests.CheckInt64(t, 2, int64(len(sa)))

	for i := range sa {
		switch sa[i].Data().Id {
		case 1:
			tests.CheckInt64(t, asc1.Calculate(), sa[i].Data().Score)
		case 2:
			tests.CheckInt64(t, asc2.Calculate(), sa[i].Data().Score)
		}
	}
}

func createUser(d data.User) (u content.User) {
	u = repo.User()
	u.Data(d)
	u.Update()

	return
}
