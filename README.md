
# Sample

Thought I'd try out gin this time.

Because it's microservices I've spent time on boilerplate and not really done much implementation, and so there's very little interesting Go code.

## sample

single binary project that can launch as any component, or all.

`go install github.com/undeconstructed/sample/sample && sample test`

test mode is config on 8001, store on 8002, frontend on 8088.

```
curl 'localhost:8088/feed?query=asdasd'

curl -X POST -v -H 'content-type: application/json' 'localhost:8002/feeds/feed1' -d '{articles:[]}'
```

## frontend

serves api - always on. could cache all the recent articles probably.

REST. could do this?

```
/feeds/
/feed/id/
/feeds/id/items/
/feeds/id/items/id
/feeds/id/items/id?in=html
```

But then again, does the app actually access the resources, or just a view of them? Frankly it could be just:

```
/feed?query=q&from=token
/article/id
```

## fetcher

gets from sources - runs whenever.

on a loop asks the config server what it should fetch, then fetches.

## store

accepts from fetcher, serves to frontend. basically just a dumb store.

```
/feeds/id
  POST many articles
  GET
/feeds/id/articles/id
  PUT article
  GET
```

## config

just config server. accepts sources, remembers them, shares the list with the frontend, gives instructions to the fetcher.
