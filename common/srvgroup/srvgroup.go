package srvgroup

import "context"

type Service interface {
	Start(context.Context)
}

type Group struct {
}

func New() *Group {
	grp := &Group{}
	return grp
}

func WithContext(context.Context) *Group {
	grp := &Group{}
	return grp
}

func (g *Group) Add(f func(context.Context) error) {
}
