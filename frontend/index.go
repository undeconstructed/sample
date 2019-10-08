package frontend

import (
	"github.com/undeconstructed/sample/common"
)

type IndexUpdater interface {
	Update(articles []common.OutputArticle)
}

type IndexQuerier interface {
	Query() []common.OutputArticle
}

type ArticleIndex struct {
	list []common.OutputArticle
}

func (a *ArticleIndex) Update(articles []common.OutputArticle) {
	a.list = articles
}

func (a *ArticleIndex) Query() []common.OutputArticle {
	list := a.list

	articles := make([]common.OutputArticle, 0, 10)

	for i := len(list) - 1; i >= 0; i-- {
		articles = append(articles, list[i])
	}

	return articles
}
