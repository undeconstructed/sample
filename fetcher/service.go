package fetcher

import (
	"context"
	"fmt"

	"github.com/mmcdole/gofeed"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	"github.com/undeconstructed/sample/common"
)

var log = logrus.WithField("service", "fetcher")

// New makes a new fetcher
func New(configURL string) common.Service {
	return &service{configURL: configURL}
}

// NewFromArgs tries to parse command line args into a service
func NewFromArgs(args []string) common.Service {
	return New(args[0])
}

type service struct {
	configURL string
}

func (s *service) Start(ctx context.Context) error {
	log.Info("Starting")

	grp, gctx := errgroup.WithContext(ctx)

	grp.Go(func() error {
		for {
			s.doFetch(gctx)
			select {
			case <-gctx.Done():
				return gctx.Err()
			default:
			}
		}
	})

	return grp.Wait()
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
		feed, err := fetchFeed(ctx, job)
		if err != nil {
			log.WithError(err).Error("Error fetching")
			continue
		}
		err = storeFeed(ctx, job, feed)
		if err != nil {
			log.WithError(err).Error("Error storing")
			continue
		}
		// TODO - ack job done
	}
}

func fetchFeed(ctx context.Context, job *common.FetchJob) ([]*common.StoreArticle, error) {
	// TODO - etag or similar
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(job.URL)

	if err != nil {
		return nil, fmt.Errorf("Fetching %w", err)
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

	return articles, nil
}

func storeFeed(ctx context.Context, job *common.FetchJob, feed []*common.StoreArticle) error {
	conn, err := grpc.Dial(job.Store, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("Dialing %w", err)
	}
	defer conn.Close()
	c := common.NewStoreClient(conn)

	toPush := &common.StorePostFeedRequest{
		FeedID:   job.ID,
		Articles: feed,
	}

	_, err = c.PostFeed(ctx, toPush)
	if err != nil {
		return fmt.Errorf("Storing %w", err)
	}

	return nil
}
