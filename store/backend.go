package store

import (
	"context"
	"fmt"
)

type backend struct {
	feeds someFeeds
}

func makeBackend() (*backend, error) {
	feeds := someFeeds{}

	return &backend{
		feeds: feeds,
	}, nil
}

func (b *backend) Start(ctx context.Context) error {
	err := b.loop(ctx)
	return err
}

func (b *backend) loop(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (b *backend) Put(ctx context.Context, fid string, article storeArticle) error {
	feed, exists := b.feeds[fid]
	if !exists {
		feed = newFeedHolder(fid)
		b.feeds[feed.id] = feed
	}

	feed.add(article)

	return nil
}

func (b *backend) Query(ctx context.Context, fid string) ([]storeArticle, error) {
	feed, exists := b.feeds[fid]
	if !exists {
		return nil, fmt.Errorf("no feed: %s", fid)
	}

	return feed.getSomeArticles(), nil
}
