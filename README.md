###zfs support for docker 1.2.0 [based on zfs_driver](https://github.com/gurjeet/docker/tree/zfs_driver/graphdriver)
copy this source in the docker source tree and compile. 

#####This are the docker  options
/zpool/docker is a example zfs dataset.<br />
DOCKER_OPTS="-g /zpool/docker -r=false -s zfs"


######that`s all
