
# Sample

Thought I'd try out gin this time.

Because it's microservices I've spent time on boilerplate and not really done much implementation, and so there's very little interesting Go code.

I'll probably carry on a bit, since it's actually not a bad project to try out some microservice ideas on.

## sample

Single binary project that can launch as any component, or all.

`go install github.com/undeconstructed/sample/sample && sample test`

test mode is config on 8001, store on 8002, frontend on 8088.

```
curl 'localhost:8088/feed?query=asdasd'

curl -X POST -v -H 'content-type: application/json' 'localhost:8002/feeds/feed1' -d '{articles:[]}'
```

## frontend

Serves REST API - always on, can run any number of them. Caches all recent articles, updating on a timer.

The API could be like this.

```
/feeds/
/feeds/id/
/feeds/id/items/
/feeds/id/items/id
/feeds/id/items/id?in=html
```

But then again, does the client actually access the resources, or just a view of them? Frankly it could be just:

```
/feed?query=q&from=token
```

Where `/feed` represents a single resource that generates a feed from internal resources.

## fetcher

Fetches from sources - runs whenever, can fail at any time without causing problems. Could also be turned into a per-task process, although that would make it harder to remember the state as of the last-fetch of a source.

Currently it asks the config server what it should fetch, then fetches, then go back to the start.

## store

Accepts from fetcher, serves to frontend - basically just a dumb store. Can fail for short periods, as the frontend is able to continue running without it.

```
/feeds/id
  POST many articles
  GET
/feeds/id/articles/id
  PUT article
  GET
```

## config

Config server. accepts sources, remembers them, shares the list with the frontend, gives instructions to the fetcher. Can fail for short periods, as the frontend is able to continue running without it.
