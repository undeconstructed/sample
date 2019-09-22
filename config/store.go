package config

import (
	"context"
	"encoding/json"
	"io/ioutil"

	"github.com/undeconstructed/sample/common"
)

type cchange struct {
	changes []interface{}
	resCh   chan error
}

type cfg struct {
	Sources map[string]common.SourceConfig `json:"sources"`
}

type store struct {
	path     string
	storeURL string

	onCh chan *cfg
	chCh chan cchange

	bytes []byte
}

func makeStore(path string, storeURL string) (*store, error) {
	return &store{
		path:     path,
		storeURL: storeURL,
		onCh:     make(chan *cfg),
		chCh:     make(chan cchange),
	}, nil
}

func (s *store) Start(ctx context.Context) error {
	bytes, cfg0, err := readConfigFile(s.path)
	if err != nil {
		log.WithError(err).Info("Using blank new config")
		cfg0 = cfg{
			Sources: map[string]common.SourceConfig{},
		}

		bytes, err = writeConfigFile(s.path, cfg0)
		if err != nil {
			return err
		}
	}

	s.bytes = bytes
	s.onCh <- &cfg0

	for {
		select {
		case c := <-s.chCh:
			err := s.write(c.changes)
			if err != nil {
				log.WithError(err).Error("Error writing config")
			}
			c.resCh <- err
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func readConfigFile(path string) (bytes []byte, cfg cfg, err error) {
	bytes, err = ioutil.ReadFile(path)
	if err != nil {
		return
	}

	err = json.Unmarshal(bytes, &cfg)
	if err != nil {
		return
	}

	return bytes, cfg, nil
}

func writeConfigFile(path string, cfg cfg) ([]byte, error) {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, err
	}

	err = ioutil.WriteFile(path, data, 0600)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *store) write(changes []interface{}) error {
	var cfg cfg
	err := json.Unmarshal(s.bytes, &cfg)
	if err != nil {
		return err
	}

	for _, change := range changes {
		switch c := change.(type) {
		case common.SourceConfig:
			if c.URL != "" {
				if c.Store == "" {
					c.Store = s.storeURL
				}
				cfg.Sources[c.ID] = c
			} else {
				delete(cfg.Sources, c.ID)
			}
		}
	}

	bytes, err := writeConfigFile(s.path, cfg)
	if err != nil {
		return err
	}
	s.bytes = bytes
	s.onCh <- &cfg
	return nil
}

func (s *store) Update(changes []interface{}) error {
	resCh := make(chan error)
	s.chCh <- cchange{
		changes: changes,
		resCh:   resCh,
	}
	return <-resCh
}
