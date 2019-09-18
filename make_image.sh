
imagename=$1

container=$(buildah from scratch)
buildah config --cmd /usr/bin/app $container
buildah config --created-by "creator" $container
buildah config --author "author at place" --label name=$imagename $container
buildah copy $container ./_build/app /usr/bin/
buildah unmount $container
image=$(buildah commit $container $imagename)
buildah rm $container

echo "made image $image"
