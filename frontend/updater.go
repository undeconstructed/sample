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

type feedCaches map[string]*feedCache

type updater struct {
	configURL string

	caches     feedCaches
	lastUpdate time.Time
	index      *ArticleIndex
}

func makeUpdater(configURL string) (*updater, error) {
	return &updater{
		configURL: configURL,
		caches:    map[string]*feedCache{},
	}, nil
}

func (s *updater) Start(ctx context.Context, index *ArticleIndex) error {
	s.index = index

	conn, err := grpc.Dial(s.configURL, grpc.WithInsecure())
	if err != nil {
		log.WithError(err).Error("Error getting source list")
		return err
	}
	defer conn.Close()
	config := common.NewConfigClient(conn)

	for {
		now := time.Now()
		s.caches = updateSources(ctx, config, s.caches)
		for _, cache := range s.caches {
			exp := now.Add(-24 * time.Hour)
			cache.articles = removeOldArticles(cache.articles, exp)
		}
		updateFeeds(s.lastUpdate, s.caches)
		s.index.Update(merge(s.caches))
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
func updateSources(ctx context.Context, config common.ConfigClient, oldCaches feedCaches) feedCaches {
	work, err := config.GetServeWork(ctx, &common.Nil{})
	if err != nil {
		log.WithError(err).Error("Error getting source list")
		return oldCaches
	}

	newCaches := map[string]*feedCache{}
	for _, feed := range work.Feeds {
		cache, exists := oldCaches[feed.ID]
		if !exists {
			cache = &feedCache{
				config: *feed,
			}
		}
		newCaches[feed.ID] = cache
	}

	return newCaches
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
func updateFeeds(since time.Time, caches feedCaches) {
	for feedID, cache := range caches {
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
