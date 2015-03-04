package test

import (
	"testing"
	"time"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/readeef/tests"
)

func TestFeed(t *testing.T) {
	f := repo.Feed()
	f.Data(data.Feed{Title: "feed title", Link: "http://sugr.org/en/products/gearshift"})

	tests.CheckInt64(t, 0, int64(f.Data().Id))

	f.Update()

	tests.CheckBool(t, false, f.HasErr(), f.Err())
	tests.CheckBool(t, false, f.Data().Id == 0)
	tests.CheckInt64(t, 0, int64(len(f.NewArticles())))

	now := time.Now()

	f.AddArticles([]content.Article{
		createArticle(data.Article{Title: "article1", Date: now, Link: "http://sugr.org/en/products/gearshift"}),
		createArticle(data.Article{Title: "article2", Date: now.Add(2 * time.Hour), Link: "http://sugr.org/en/products/readeef"}),
		createArticle(data.Article{Title: "article3", Date: now.Add(-3 * time.Hour), Link: "http://sugr.org/en/about/us"}),
	})
	tests.CheckBool(t, false, f.HasErr(), f.Err())

	tests.CheckInt64(t, 3, int64(len(f.NewArticles())))

	f.AddArticles([]content.Article{
		createArticle(data.Article{Title: "article4", Date: now.Add(-10 * 24 * time.Hour), Link: "http://sugr.org/bg/"}),
	})
	tests.CheckBool(t, false, f.HasErr(), f.Err())

	tests.CheckInt64(t, 1, int64(len(f.NewArticles())))
	tests.CheckString(t, "article4", f.NewArticles()[0].Data().Title)

	a := f.AllArticles()

	tests.CheckBool(t, false, f.HasErr(), f.Err())
	tests.CheckInt64(t, 4, int64(len(a)))

	for i := range a {
		d := a[i].Data()
		switch d.Title {
		case "article1":
		case "article2":
		case "article3":
		case "article4":
		default:
			tests.CheckBool(t, false, true, "Unknown article")
		}
	}

	a = f.LatestArticles()
	tests.CheckBool(t, false, f.HasErr(), f.Err())
	tests.CheckInt64(t, 3, int64(len(a)))

	for i := range a {
		d := a[i].Data()
		switch d.Title {
		case "article1":
		case "article2":
		case "article3":
		default:
			tests.CheckBool(t, false, true, "Unknown article")
		}
	}

}

func TestUserFeed(t *testing.T) {
	uf := repo.UserFeed(createUser(data.User{}))
	uf.Data(data.Feed{Link: "http://sugr.org"})

	tests.CheckBool(t, false, uf.Validate() == nil)

	u := createUser(data.User{Login: "user_feed_login"})

	uf = repo.UserFeed(u)
	uf.Data(data.Feed{Link: "http://sugr.org", Title: "User feed 1"})

	tests.CheckBool(t, true, uf.Validate() == nil, uf.Validate())

	uf.Update()
	tests.CheckBool(t, false, uf.HasErr(), uf.Err())

	u.AddFeed(uf)

	id := uf.Data().Id

	uf2 := u.FeedById(id)
	tests.CheckBool(t, false, uf2.HasErr(), uf2.Err())
	tests.CheckString(t, uf.Data().Title, uf2.Data().Title)

	now := time.Now()

	uf.AddArticles([]content.Article{
		createArticle(data.Article{Id: 1, Title: "article1", Date: now, Link: "http://sugr.org/en/products/gearshift"}),
		createArticle(data.Article{Id: 2, Title: "article2", Date: now.Add(2 * time.Hour), Link: "http://sugr.org/en/products/readeef"}),
		createArticle(data.Article{Id: 3, Title: "article3", Date: now.Add(-3 * time.Hour), Link: "http://sugr.org/en/about/us"}),
	})
	tests.CheckBool(t, false, uf.HasErr(), uf.Err())

	ua := uf.Articles()
	tests.CheckBool(t, false, uf.HasErr(), uf.Err())

	tests.CheckInt64(t, 3, int64(len(ua)))

	tests.CheckInt64(t, 1, int64(ua[0].Data().Id))
	tests.CheckString(t, "article2", ua[1].Data().Title)
	tests.CheckBool(t, true, now.Add(-3*time.Hour).Equal(ua[2].Data().Date))

	uf.SortingByDate()
	ua = uf.Articles()

	tests.CheckInt64(t, 3, int64(ua[0].Data().Id))
	tests.CheckString(t, "article1", ua[1].Data().Title)
	tests.CheckBool(t, true, now.Add(2*time.Hour).Equal(ua[2].Data().Date))

	uf.Reverse()
	ua = uf.Articles()

	tests.CheckInt64(t, 2, int64(ua[0].Data().Id))
	tests.CheckString(t, "article1", ua[1].Data().Title)
	tests.CheckBool(t, true, now.Add(-3*time.Hour).Equal(ua[2].Data().Date))

	ua[0].Read(true)

	uf.Reverse()
	uf.DefaultSorting()

	ua = uf.UnreadArticles()
	tests.CheckBool(t, false, uf.HasErr(), uf.Err())
	tests.CheckInt64(t, 2, int64(len(ua)))

	tests.CheckInt64(t, 1, int64(ua[0].Data().Id))
	tests.CheckString(t, "article3", ua[1].Data().Title)

	u.ArticleById(data.ArticleId(2)).Read(false)

	ua = uf.UnreadArticles()
	tests.CheckInt64(t, 3, int64(len(ua)))

	uf.ReadBefore(now.Add(time.Minute), true)
	tests.CheckBool(t, false, uf.HasErr(), uf.Err())

	ua = uf.UnreadArticles()
	tests.CheckBool(t, false, uf.HasErr(), uf.Err())
	tests.CheckInt64(t, 1, int64(len(ua)))
	tests.CheckInt64(t, 2, int64(ua[0].Data().Id))

	uf.Detach()
	tests.CheckInt64(t, 0, int64(len(u.AllFeeds())))

	uf2 = u.FeedById(id)
	tests.CheckBool(t, true, uf2.Err() == content.ErrNoContent)

	asc1 := createArticleScores(data.ArticleScores{ArticleId: 1, Score1: 2, Score2: 2})
	asc2 := createArticleScores(data.ArticleScores{ArticleId: 2, Score1: 1, Score2: 3})

	sa := uf.ScoredArticles(now.Add(-20*time.Hour), now.Add(20*time.Hour))

	tests.CheckBool(t, false, uf.HasErr(), uf.Err())
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

func createFeed(d data.Feed) (f content.Feed) {
	f = repo.Feed()
	f.Data(d)
	f.Update()

	return
}

func createUserFeed(u content.User, d data.Feed) (uf content.UserFeed) {
	uf = repo.UserFeed(u)
	uf.Data(d)
	uf.Update()

	u.AddFeed(uf)

	return
}

func createTaggedFeed(u content.User, d data.Feed) (tf content.TaggedFeed) {
	tf = repo.TaggedFeed(u)
	tf.Data(d)
	tf.Update()

	u.AddFeed(tf)

	return
}
