package fetcher

import (
	"context"

	"github.com/mmcdole/gofeed"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/undeconstructed/sample/common"
)

var log = logrus.WithField("service", "fetcher")

// New makes a new fetcher
func New(configURL string) common.Service {
	return &service{configURL: configURL}
}

type service struct {
	configURL string

	stop   context.CancelFunc
	config common.ConfigClient
}

func (s *service) Start() error {
	log.Info("Starting")

	ctx, cancel := context.WithCancel(context.Background())
	s.stop = cancel

	go func() {
		for {
			s.doFetch(ctx)
			select {
			case <-ctx.Done():
				return
			default:
			}
		}
	}()

	return nil
}

func (s *service) doFetch(ctx context.Context) {
	conn, err := grpc.Dial(s.configURL, grpc.WithInsecure())
	if err != nil {
		log.WithError(err).Error("Error connecting to config")
		return
	}
	defer conn.Close()
	c := common.NewConfigClient(conn)

	work, err := c.GetWork(ctx, &common.Nil{})
	if err != nil {
		log.WithError(err).Error("Error getting work list")
		return
	}

	for _, job := range work.Jobs {
		log.WithField("job", job).Info("Fetching")
		s.fetchFeed(ctx, job)
	}
}

func (s *service) fetchFeed(ctx context.Context, job *common.FetchJob) {
	// TODO - etag or similar
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(job.URL)

	if err != nil {
		log.WithError(err).Error("Error fetching")
		return
	}

	articles := make([]*common.StoreArticle, 0, len(feed.Items))
	for _, item := range feed.Items {
		articles = append(articles, &common.StoreArticle{
			ID:    item.GUID,
			Title: item.Title,
			Date:  item.PublishedParsed.Unix(),
			Body:  item.Content,
		})
	}

	conn, err := grpc.Dial(job.Store, grpc.WithInsecure())
	if err != nil {
		log.WithError(err).Error("Error connecting to store")
		return
	}
	defer conn.Close()
	c := common.NewStoreClient(conn)

	toPush := &common.StorePostFeedRequest{
		FeedID:   job.ID,
		Articles: articles,
	}

	_, err = c.PostFeed(ctx, toPush)
	if err != nil {
		log.WithError(err).Error("Error storing articles")
	}

	// TODO - ack job done
}

func (s *service) Stop() error {
	s.stop()
	return nil
}
