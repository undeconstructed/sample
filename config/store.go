package config

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	"github.com/undeconstructed/sample/common"
)

type sf func(*cfg)

type cfg struct {
	Sources map[string]common.SourceConfig `json:"sources"`
}

type store struct {
	cfg
	sync.RWMutex
}

func makeStore() *store {
	cfg0, err := readConfigFile()
	if err != nil {
		log.WithError(err).Info("Using blank new config")
		cfg0 = cfg{
			Sources: map[string]common.SourceConfig{},
		}
	}

	return &store{
		cfg: cfg0,
	}
}

func readConfigFile() (cfg cfg, err error) {
	data, err := ioutil.ReadFile("config.json")
	if err != nil {
		return
	}

	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return
	}

	return cfg, nil
}

func writeConfigFile(cfg cfg) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile("config.json", data, 0600)
	if err != nil {
		return err
	}

	return nil
}

func (s *store) read(f sf) {
	s.RLock()
	defer s.RUnlock()
	f(&s.cfg)
}

func (s *store) write(f sf) {
	s.Lock()
	defer s.Unlock()
	f(&s.cfg)
	err := writeConfigFile(s.cfg)
	if err != nil {
		log.WithError(err).Error("config not saved, evertyhing terrible")
		return
	}
	log.Info("config changed")
}
