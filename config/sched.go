package config

import (
	"context"
	"time"

	"github.com/undeconstructed/sample/common"
)

type schedReq struct {
	resCh chan common.FetchWork
}

type schedEntry struct {
	source common.SourceConfig
	time   int64
}

type sched struct {
	reqCh chan schedReq
	cfgCh chan *cfg

	table map[string]*schedEntry
}

func makeSched() *sched {
	reqCh := make(chan schedReq, 10)
	cfgCh := make(chan *cfg)

	return &sched{
		reqCh: reqCh,
		cfgCh: cfgCh,
		table: map[string]*schedEntry{},
	}
}

func (s *sched) start(ctx context.Context) error {
	for {
		now := time.Now().Unix()
		cutoff := now - 60
		toDo := s.findNext(cutoff)

		if toDo == nil {
			t := time.After(60 * time.Second)
			select {
			case <-t:
				continue
			case c := <-s.cfgCh:
				s.updateSources(c.Sources)
			case <-ctx.Done():
				return ctx.Err()
			}
		} else {
			select {
			case r := <-s.reqCh:
				toDo.time = now
				r.resCh <- common.FetchWork{
					Jobs: []*common.FetchJob{
						&common.FetchJob{
							ID:    toDo.source.ID,
							URL:   toDo.source.URL,
							Store: toDo.source.Store,
						},
					},
				}
			case c := <-s.cfgCh:
				s.updateSources(c.Sources)
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}

func (s *sched) updateSources(sources map[string]common.SourceConfig) {
	table := map[string]*schedEntry{}
	for _, source := range sources {
		old, exists := s.table[source.ID]
		if !exists {
			table[source.ID] = &schedEntry{source: source}
		} else {
			// TODO - changed entry
			table[source.ID] = old
		}
	}
	s.table = table
}

func (s *sched) findNext(cutoff int64) *schedEntry {
	for _, e := range s.table {
		if e.time < cutoff {
			return e
		}
	}
	return nil
}

func (s *sched) getWork() common.FetchWork {
	resCh := make(chan common.FetchWork)
	s.reqCh <- schedReq{resCh}
	return <-resCh
}
