
frontend
  serves api - always on
fetcher
  gets from sources - runs whenever
store
  accepts from fetcher, serves to frontend
config
  just config server

no users. no state.

REST. sigh. could do this?

/feeds/
/feed/id/
/feeds/id/items/
/feeds/id/items/id
/feeds/id/items/id?in=html

But then again, does the app actually access the resources, or just a view of them? Frankly it could be just:

/feed?q=q
/article/id
