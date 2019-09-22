
# Sample

Learning project for some microservice on Kubernetes things.

## Architecture

* 1 config service
* 1+ stores to hold fetched data
* 0+ fetchers to fetch from remote sources and push to stores
* 1+ frontends to serve HTTP to clients

All services communicate with gRPC. Config has additional HTTP server for easy control.

## Useful commands

Full build:

```
make app image
```

Run an local image repo:

```
podman run -p 5000:5000 \
  docker.io/library/registry
```

Push image to local repo:

```
podman push --tls-verify=0 \
  localhost/sample-1:latest \
  docker://localhost:5000/sample-1:v1.0.0-1
```

Create and start a minikube, allowing access to local image repo:

```
minikube start \
  --vm-driver=kvm2 \
  --network-plugin=cni \
  --enable-default-cni \
  --container-runtime=cri-o \
  --insecure-registry=192.168.122.1:5000,minikube.host:5000 \
  --bootstrapper=kubeadm

```

Apply all k8s config (e.g. to the minikube):

```
kubectl apply -f k8s/
```

Get the virtual service URLs from minikube:

```
CONFIG=$(minikube service --url=true config-public
PUBLIC_SERVICE=$(minikube service --url=true public-service)
```

Make some config:

```
curl -XPUT -H"content-type: application/json" "$CONFIG/sources/bbc" \
  -d '{"url":"http://feeds.bbci.co.uk/news/uk/rss.xml","store":"store-0.store:8000"}'

curl -XGET "$CONFIG/sources"
```

Try the public endpoint:

```
curl "$PUBLIC_SERVICE/feed?query=asdasd"
```

Or browser:

```
xdg-open "$CONFIG/sources"
xdg-open "$PUBLIC_SERVICE/feed"
```

## Local/test mode

```
go run github.com/undeconstructed/sample/sample test
```

test mode is config http on 8087, frontend http on 8088.

```
curl -XPUT -H"content-type: application/json" localhost:8087/sources/bbc -d '{"url":"http://feeds.bbci.co.uk/news/uk/rss.xml"}'

curl 'localhost:8088/feed?query=asdasd'
```

## Directories

### k8s

Kubernetes config yaml.

### common

Useful things, shared types, gRPC specs.

### sample

Single binary project that can launch as any component, or all.

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

But then again, does the client actually access the resources, or just a view of them? So in fact it's just:

```
/feed?query=q&from=token
```

Where `/feed` represents a single resource that generates a feed from internal resources.

## config

Config server. accepts sources, remembers them, shares the list with the frontend, gives instructions to the fetcher. Can fail for short periods, as the frontend is able to continue running without it.

## fetcher

Fetches from sources - runs whenever, can fail at any time without causing problems. Could also be turned into a per-task process, although that would make it harder to remember the state as of the last-fetch of a source.

Asks the config server what it should fetch, then fetches, then goes back to the start.

## store

Accepts from fetcher, serves to frontend - basically just a dumb store. Can fail for short periods, as the frontend is able to continue running without it.

Embeds a database to provide its own persistence.
