package frontend

import (
	"github.com/undeconstructed/sample/common"
)

type ArticleIndex struct {
	list []common.OutputArticle
}

func (a *ArticleIndex) Update(articles []common.OutputArticle) {
	a.list = articles
}

func (a *ArticleIndex) Query() []common.OutputArticle {
	articles := make([]common.OutputArticle, 0, 10)

	for i := len(a.list) - 1; i >= 0; i-- {
		articles = append(articles, a.list[i])
	}

	return articles
}
