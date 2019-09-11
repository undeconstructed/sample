package frontend

import (
	"context"
	"time"

	"google.golang.org/grpc"

	"github.com/undeconstructed/sample/common"
)

type updater struct {
	configURL string

	sources    []*common.ConfigSource
	lastUpdate time.Time
	articles   someArticles
}

func makeUpdater(configURL string, articles someArticles) (*updater, error) {
	return &updater{
		configURL: configURL,
		articles:  articles,
	}, nil
}

func (s *updater) start(ctx context.Context) error {
	// could update on request instead, of course
	for {
		now := time.Now()
		s.updateSources()
		s.purgeArticles(now)
		s.updateFeeds(s.lastUpdate)
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

	sources, err := c.GetSources(context.Background(), &common.Nil{})
	if err != nil {
		log.WithError(err).Error("Error getting source list")
		return
	}

	// atomic replace of whole list
	s.sources = sources.Sources
}

// get rid of old things
func (s *updater) purgeArticles(before time.Time) {
}

// get any new articles from the store
func (s *updater) updateFeeds(since time.Time) {
	for _, source := range s.sources {

		conn, err := grpc.Dial(source.Store, grpc.WithInsecure())
		if err != nil {
			log.WithError(err).Error("Error connecting to store")
			return
		}
		defer conn.Close()
		c := common.NewStoreClient(conn)

		req := &common.StoreGetFeedRequest{
			FeedID: source.ID,
		}

		res, err := c.GetFeed(context.Background(), req)
		if err != nil {
			log.WithError(err).Error("Error updating feed")
			return
		}

		newArticles := make([]common.OutputArticle, 0, len(res.Articles))
		for _, a := range res.Articles {
			newArticles = append(newArticles, common.OutputArticle{
				Source: a.ID,
				ID:     a.ID,
				Title:  a.Title,
				Date:   time.Unix(a.Date, 0),
				Body:   a.Body,
			})
		}

		// TODO - merge, not replace
		s.articles.list = newArticles
	}
}
