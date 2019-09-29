package frontend

import (
	"context"
	"time"

	"google.golang.org/grpc"

	"github.com/undeconstructed/sample/common"
)

type feedCache struct {
	config   common.ServeFeed
	articles []common.OutputArticle
}

type updater struct {
	configURL string

	caches     map[string]*feedCache
	lastUpdate time.Time
	articles   someArticles
}

func makeUpdater(configURL string, articles someArticles) (*updater, error) {
	return &updater{
		configURL: configURL,
		caches:    map[string]*feedCache{},
		articles:  articles,
	}, nil
}

func (s *updater) Start(ctx context.Context) error {
	// could update on request instead, of course
	for {
		now := time.Now()
		s.updateSources()
		s.purgeArticles(now)
		s.updateFeeds(s.lastUpdate)
		s.articles.list = merge(s.caches)
		s.lastUpdate = now
		t := time.After(10 * time.Second)
		select {
		case <-t:
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// find out what sources exist
func (s *updater) updateSources() {
	conn, err := grpc.Dial(s.configURL, grpc.WithInsecure())
	if err != nil {
		log.WithError(err).Error("Error getting source list")
		return
	}
	defer conn.Close()
	c := common.NewConfigClient(conn)

	work, err := c.GetServeWork(context.Background(), &common.Nil{})
	if err != nil {
		log.WithError(err).Error("Error getting source list")
		return
	}

	newCaches := map[string]*feedCache{}
	for _, feed := range work.Feeds {
		cache, exists := s.caches[feed.ID]
		if !exists {
			cache = &feedCache{
				config: *feed,
			}
		}
		newCaches[feed.ID] = cache
	}

	s.caches = newCaches
}

// get rid of old things
func (s *updater) purgeArticles(before time.Time) {
	for _, cache := range s.caches {
		cache.articles = removeOldArticles(cache.articles, before)
	}
}

func removeOldArticles(list []common.OutputArticle, before time.Time) []common.OutputArticle {
	i := 0
	for _, article := range list {
		if !article.Date.Before(before) {
			break
		}
		i++
	}
	newArticles := make([]common.OutputArticle, len(list)-i)
	copy(newArticles, list[i:])
	return newArticles
}

// get any new articles from the store
func (s *updater) updateFeeds(since time.Time) {
	for feedID, cache := range s.caches {

		conn, err := grpc.Dial(cache.config.Store, grpc.WithInsecure())
		if err != nil {
			log.WithError(err).Error("Error connecting to store")
			return
		}
		defer conn.Close()
		c := common.NewStoreClient(conn)

		req := &common.StoreGetFeedRequest{
			FeedID: feedID,
		}

		res, err := c.GetFeed(context.Background(), req)
		if err != nil {
			log.WithError(err).Error("Error updating feed")
			return
		}

		newArticles := make([]common.OutputArticle, 0, len(res.Articles))
		for _, a := range res.Articles {
			newArticles = append(newArticles, common.OutputArticle{
				Source: feedID,
				ID:     a.ID,
				Title:  a.Title,
				Date:   time.Unix(a.Date, 0),
				Body:   a.Body,
			})
		}

		cache.articles = newArticles
	}
}

func merge(caches map[string]*feedCache) []common.OutputArticle {
	inputs := [][]common.OutputArticle{}
	for _, cache := range caches {
		inputs = append(inputs, cache.articles)
	}

	newArticles := []common.OutputArticle{}
	for {
		next := -1
		var nextT *time.Time
		for i, input := range inputs {
			if len(input) > 0 {
				t := input[0].Date
				if nextT == nil || t.Before(*nextT) {
					nextT = &t
					next = i
				}
			}
		}
		if next == -1 {
			break
		}
		source := inputs[next]
		newArticles = append(newArticles, source[0])
		inputs[next] = source[1:]
	}

	return newArticles
}
