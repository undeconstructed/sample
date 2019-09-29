package frontend

import (
	"testing"
	"time"

	"github.com/undeconstructed/sample/common"
)

func TestMerge(t *testing.T) {
	i := map[string]*sourceCache{}
	i["s1"] = &sourceCache{
		articles: []common.OutputArticle{
			{
				Source: "s1",
				ID:     "2",
				Date:   time.Unix(2, 0),
			},
			{
				Source: "s1",
				ID:     "3",
				Date:   time.Unix(3, 0),
			},
		},
	}
	i["s2"] = &sourceCache{
		articles: []common.OutputArticle{
			{
				Source: "s2",
				ID:     "1",
				Date:   time.Unix(1, 0),
			},
			{
				Source: "s2",
				ID:     "4",
				Date:   time.Unix(4, 0),
			},
		},
	}
	r := merge(i)
	r0 := ""
	for _, a := range r {
		r0 += a.ID
	}
	if r0 != "1234" {
		t.Fail()
	}
}

func TestRemoveOldArticles(t *testing.T) {
	i := []common.OutputArticle{
		{
			Source: "s1",
			ID:     "2",
			Date:   time.Unix(2, 0),
		},
		{
			Source: "s1",
			ID:     "3",
			Date:   time.Unix(3, 0),
		},
		{
			Source: "s1",
			ID:     "4",
			Date:   time.Unix(3, 0),
		},
	}
	r := removeOldArticles(i, time.Unix(3, 0))
	r0 := ""
	for _, a := range r {
		r0 += a.ID
	}
	if r0 != "34" {
		t.Errorf("test: %v\n", r0)
		t.Fail()
	}
}
