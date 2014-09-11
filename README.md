###zfs support for docker 1.2.0 [based on zfs_driver](https://github.com/gurjeet/docker/tree/zfs_driver/graphdriver)
copy this in the docker source tree and compile. 

#####This are the options
/zpool/docker ist zfs dataset.
DOCKER_OPTS="-g /zpool/docker -r=false -s zfs"


######that`s all
