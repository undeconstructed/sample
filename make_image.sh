
imagename=$1

container=$(buildah from scratch)
buildah config \
  --entrypoint '[ "/usr/bin/app" ]' \
  --created-by "a script" \
  --author "Phil <address>" \
  --label name=$imagename \
  $container
buildah copy $container ./_build/app /usr/bin/
buildah unmount $container
image=$(buildah commit $container $imagename)
buildah rm $container

echo "made image $image"
