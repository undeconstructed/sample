package store

import (
	"context"
	"encoding/binary"
	"encoding/json"

	bolt "go.etcd.io/bbolt"
)

// storeArticle is an article as stored in the store.
type storeArticle struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Date  int64  `json:"date"`
	Body  string `json:"body"`
}

type backend struct {
	path string

	db *bolt.DB
}

func makeBackend(path string) (*backend, error) {
	return &backend{
		path: path,
	}, nil
}

func (b *backend) Start(ctx context.Context) error {
	db, err := bolt.Open(b.path, 0666, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	b.db = db

	return b.loop(ctx)
}

func (b *backend) loop(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (b *backend) Put(ctx context.Context, fid string, articles []storeArticle) error {
	bo := binary.BigEndian

	err := b.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(fid))
		if err != nil {
			return err
		}

		for _, a := range articles {
			k := make([]byte, 8)
			bo.PutUint64(k, uint64(a.Date))

			// XXX - no need to to this inside txn
			v, err := json.Marshal(a)
			if err != nil {
				// scrap whole thing
				return err
			}

			err = b.Put(k, v)
			if err != nil {
				// scrap whole thing
				return err
			}
		}

		return nil
	})

	return err
}

func (b *backend) Query(ctx context.Context, fid string) ([]storeArticle, error) {
	var out []storeArticle

	addArticle := func(data []byte) error {
		a := storeArticle{}
		err := json.Unmarshal(data, &a)
		if err != nil {
			return err
		}
		out = append(out, a)
		return nil
	}

	err := b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(fid))
		if b == nil {
			return nil
		}
		c := b.Cursor()
		n := 0
		for k, v := c.First(); k != nil && n < 10; k, v = c.Next() {
			// XXX - no need to to this inside txn
			err := addArticle(v)
			if err != nil {
				return err
			}
			n++
		}
		return nil
	})
	if err != nil {
		return out, err
	}

	return out, nil
}
