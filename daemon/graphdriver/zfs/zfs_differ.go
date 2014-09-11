package zfs

/*
 * NOTE: To be able to diff a ZFS filesystem against its 'origin' snapshot, one
 * must also have the origin FS of the snapshot in mounted state.
 * For e.g.
 *
 * FS				abc
 * SNAPSHOT				abc@snap_name
 * CLONE						xyz
 *
 * The abc filesystem should be in mounted state for performing
 * 	zfs diff abc@snap_name xyz
 */

/*
 * Differ interface implementation for ZFS driver.
 */

/* AUFS code follows

/*
import (
	"bufio"
	"fmt"
	"github.com/docker/docker/archive"
	"github.com/docker/docker/daemon/graphdriver"
	"github.com/docker/docker/utils"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

// Return an archive of the contents for the id
func (a *Driver) Diff(id string) (archive.Archive, error) {
	return archive.TarFilter(path.Join(a.rootPath(), "diff", id), &archive.TarOptions{
		Recursive:   true,
		Compression: archive.Uncompressed,
	})
}

func (a *Driver) Changes(id string) ([]archive.Change, error) {
	layers, err := a.getParentLayerPaths(id)
	if err != nil {
		return nil, err
	}
	return archive.Changes(layers, path.Join(a.rootPath(), "diff", id))
}

func (a *Driver) ApplyDiff(id string, diff archive.Archive) error {
	return archive.Untar(diff, path.Join(a.rootPath(), "diff", id), nil)
}

// Returns the size of the contents for the id
func (a *Driver) DiffSize(id string) (int64, error) {
	return utils.TreeSize(path.Join(a.rootPath(), "diff", id))
}

*/
