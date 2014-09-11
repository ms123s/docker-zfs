package zfs

/*
 * This file contains the public interface of the ZFS driver
 */

import (
	//"errors"
	"fmt"
	"github.com/docker/docker/daemon/graphdriver"
	"math/rand"
	"os"
	"strings"
	"time"
)

func init() {
	dbg("ZFS init") // This debug line shows up in `docker run` output, when the container is launched.

	graphdriver.Register("zfs", Init)
}

type Driver struct {
	root            string // Path to the root of the graph storage (as seen by Docker daemon)
	root_dataset    string // Name of the ZFS dataset mount at 'root'
	root_mountpoint string // Filesystem mountpoint; must be the same as 'root'
	rand            *rand.Rand
}

/*
 * Initialize the driver.
 *
 * An error is returned if ZFS is not available on the system.
 */
func Init(root string, options []string) (graphdriver.Driver, error) {
	{funcName := funcEnter();defer funcLeave(funcName)}
	dbg("ZFSInit: root=%s", root)

	// Check if the ZFS filesystem is present
	if err := supportsZFS(); err != nil {
		dbg("ZFS is not supported")
		return nil, err
	}

	dbg("ZFS is supported")

	dbg("root: %s", root)

	/*
	 * Check that the root path provided to us is a ZFS filesystem. Instruct the
	 * command to emit machine-readable output (-H) by leaving out the header and
	 * using TAB to separate the fields. `zfs create` disallows a TAB character in
	 * dataset's name, so there's no danger of us getting the mount-point wrong.
	 */
	outStream, _, err := execCmd("zfs", "list", "-H", "-o", "name,mountpoint", "-t", "filesystem", root)
	if err != nil {
		return nil, err // XXX We should cook a errors.New() with accurate message.
	}

	/*
	 * Split the output on tab characters.
	 */
	outSplice := strings.FieldsFunc(outStream,
		func(r rune) bool {
			return r == '\t'
		})

	/* No need to trim the trailing newline; execCmd() already did that. */
	dataset_name := outSplice[0]
	mount_point := outSplice[1]

	driver := Driver{root, dataset_name, mount_point, rand.New(rand.NewSource(time.Now().UnixNano()))}
	dbg("New driver object: %v", driver.Status())

	/*
	 * Now change to the directory that is the mount-point of this filesystem. The
	 * whole point of this exercise is to ensure that the filesystem can't be
	 * unmounted behind our back while we are running. The Docker daemon should not
	 * change its directory past this point, or else we lose this protection.
	 */
	if err := os.Chdir(mount_point); err != nil {
		return nil, fmt.Errorf("zfs-Init: Could not change to the mount point '%s'", mount_point)
	}

	return &driver, nil
}

func (d *Driver) String() string {
	{funcName := funcEnter();defer funcLeave(funcName)}

	return "zfs"
}

func (d *Driver) Status() [][2]string {
	{funcName := funcEnter();defer funcLeave(funcName)}

	return [][2]string{
		{"Root Dir", d.root},
		{"Dataset", d.root_dataset},
		{"Mount Point", d.root_mountpoint},
		// TODO: Emulate AUFS driver-like output; not necessary, but see what more info can help the user.
	}
}

/*
 * This is called when unmounting the driver. The driver is supposed to unmount the
 * filesystems of all the containers that it has in its registry.
 */
func (d *Driver) Cleanup() error {
	{funcName := funcEnter();defer funcLeave(funcName)}
	dbg("ZFSCleanup")

	msg := "zfs-Cleanup: not yet implemented"
	dbg(msg)
	panic(msg)
	return nil;//errors.New(msg)
}

/*
 * Create the on-disk structures for  the container's storage. Use the parent's
 * storage contents to populate the base image of this container.
 */
func (d *Driver) Create(id string, parent string) error {
	{funcName := funcEnter();defer funcLeave(funcName)}
	dbg("ZFSCreate: id=%s,parent=%s", id,parent)

	dataset := d.getDataset(id)

	/* TODO: What should we do if the container storage already exists? During
	 * development, when image creation was interrupted midway, the ZFS dataset was
	 * not cleaned up, and caused error in the next run, until dataset was manually
	 * removed.
	 */

	/*
	 * We create the new filesystem with canmount=noauto property, so that the ZFS
	 * filesystem does not try to mount these during FS creation here, and nor at
	 * OS bootup time. Apart from the possible performance implications during
	 * bootup, this is done primarily to avoid any error messages during OS bootup
	 * which may be caused by either not-yet-attached devices, or dependency of ZFS
	 * mount points on non-ZFS mount points.
	 *
	 * These filesystems are mounted, when necessary, by the Get() call.
	 */
	if parent != "" {
		snapshotName := "docker_" + fmt.Sprintf("%x", d.rand.Int31())
		snapshotPath := d.getDataset(parent) + "@" + snapshotName

		/* Create a snapshot of parent's storage */
		_, _, err := execCmd("zfs", "snapshot", snapshotPath)
		if err != nil {
			return err // XXX We should cook a errors.New() with accurate message.
		}

		/* Clone the snapshot. */
		_, _, err = execCmd("zfs", "clone", "-o", "canmount=noauto", snapshotPath, dataset)
		if err != nil {
			return err // XXX We should cook a errors.New() with accurate message.
		}

		/* Mark the snapshot to be deleted, once it's not needed anymore */
		_, _, err = execCmd("zfs", "destroy", "-d", snapshotPath)
		if err != nil {
			return err // XXX We should cook a errors.New() with accurate message.
		}

	} else {
		_, _, err := execCmd("zfs", "create", "-o", "canmount=noauto", dataset)
		if err != nil {
			return err // XXX We should cook a errors.New() with accurate message.
		}
	}

	return nil
}

/*
 * Remove the on-disk structures of the container's storage.
 */
func (d *Driver) Remove(id string) error {
	{funcName := funcEnter();defer funcLeave(funcName)}
	dbg("ZFSRemove: id=%s", id)

	dataset := d.getDataset(id)

	/*
	 * Promote a clone, if any, of this dataset, and then remove the dataset. Other
	 * snapshots and clones, if any, will automatically adjust to be dependent on
	 * this promoted clone.
	 *
	 * There seems to be no need for us to promote a clone. I did some tests, and I
	 * saw that Docker itself calls 'destroy' on container/image tree recursively,
	 * and that it makes sure that it doesn't call 'destroy' on a filesystem until
	 * all its children have been destroyed first. So as the code stands right now,
	 * it is usable by Docker.
	 *
	 * But there may be some value to promoting a clone. For eg., if a user manually
	 * creates a clone of a container's ZFS backing dataset, then upon container
	 * deletion by Docker, we do `zfs destroy -R` and that will destroy the user's
	 * manually-created snapshots and clones. In that case, promoting one of
	 * user-created clones seems to be a safe choice. Since the code is already in
	 * place, I'll just keep it around till final merge and see what Docker-core
	 * team thinks of it.
	 */
	outStream, _, err := execCmd("zfs", "list", "-r", "-H", "-o", "clones", "-t", "snapshot", dataset)
	if err != nil {
		return err // XXX We should cook a errors.New() with accurate message.
	}

	/*
	 * Split the output on comma and newline characters.
	 */
	outSplice := strings.FieldsFunc(outStream,
		func(r rune) bool {
			return r == ',' || r == '\n'
		})

	if len(outSplice) != 0 {

		clone_to_promote := outSplice[0]
		_, _, err = execCmd("zfs", "promote", clone_to_promote)
		if err != nil {
			return err // XXX We should cook a errors.New() with accurate message.
		}
	}

	/* Now destroy the container's FS, including any snapshots */
	_, _, err = execCmd("zfs", "destroy", "-r", dataset)
	if err != nil {
		return err // XXX We should cook a errors.New() with accurate message.
	}

	return nil
}

/*
 * Mount the storage of the container, and return the resulting (read-write capable)
 * path to it.
 */
func (d *Driver) Get(id string, mountLabel string) (string, error) {
	{funcName := funcEnter();defer funcLeave(funcName)}
	dbg("ZFSGet: id=%s,mountLabel:%s", id,mountLabel)

	dataset := d.getDataset(id)

	outStream, _, err := execCmd("zfs", "list", "-H", "-o", "mounted", dataset)
	if err != nil {
		return "", err // XXX We should cook a errors.New() with accurate message.
	}

	/* Return early if already mounted */
	if outStream == "yes" {
		return d.getPath(id), nil
	}

	_, _, err = execCmd("zfs", "mount", dataset)
	if err != nil {
		return "", err // XXX We should cook a errors.New() with accurate message.
	}

	return d.getPath(id), nil
}

func (d *Driver) Put(id string) {
	dbg("ZFSPut: id=%s", id)
	// Get() creates no runtime resources (like e.g. mounts)
	// so this doesn't need to do anything.
}


/*
 * Exists returns true if the given id is registered with this driver.
 */
func (d *Driver) Exists(id string) bool {
	{funcName := funcEnter();defer funcLeave(funcName)}
	dbg("ZFSExists: id=%s", id)

	dataset := d.getDataset(id)

	_, _, err := execCmd("zfs", "list", "-H", "-o", "mounted", dataset)
	if err != nil {
		return false // XXX We should cook a errors.New() with accurate message.
	}
	return true
}
