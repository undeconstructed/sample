package frontend

import (
	"github.com/undeconstructed/sample/common"
)

type someArticlesX struct {
	list []common.OutputArticle
}

type someArticles *someArticlesX
