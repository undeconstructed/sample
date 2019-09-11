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

type schedTable map[string]*schedEntry

type sched struct {
	reqCh chan schedReq
	cfgCh chan *cfg

	table schedTable
}

func makeSched() (*sched, error) {
	reqCh := make(chan schedReq)
	cfgCh := make(chan *cfg)

	return &sched{
		reqCh: reqCh,
		cfgCh: cfgCh,
		table: schedTable{},
	}, nil
}

func (s *sched) Start(ctx context.Context) error {
	err := s.loop(ctx)
	return err
}

func (s *sched) loop(ctx context.Context) error {
	for {
		now := time.Now().Unix()
		cutoff := now - 60
		toDo := findNext(s.table, cutoff)

		if toDo == nil {
			t := time.After(60 * time.Second)
			select {
			case <-t:
				continue
			case c := <-s.cfgCh:
				s.table = updateTable(s.table, c.Sources)
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
				s.table = updateTable(s.table, c.Sources)
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}

func (s *sched) getWork(ctx context.Context) (common.FetchWork, error) {
	resCh := make(chan common.FetchWork)
	// well, this is horrible now
	select {
	case s.reqCh <- schedReq{resCh}:
	case <-ctx.Done():
		return common.FetchWork{}, ctx.Err()
	}
	select {
	case r := <-resCh:
		return r, nil
	case <-ctx.Done():
		return common.FetchWork{}, ctx.Err()
	}
}

func updateTable(table schedTable, sources map[string]common.SourceConfig) schedTable {
	ntable := schedTable{}
	for _, source := range sources {
		old, exists := table[source.ID]
		if !exists {
			ntable[source.ID] = &schedEntry{source: source}
		} else {
			// TODO - changed entry
			ntable[source.ID] = old
		}
	}
	return ntable
}

func findNext(table schedTable, cutoff int64) *schedEntry {
	for _, e := range table {
		if e.time < cutoff {
			return e
		}
	}
	return nil
}
