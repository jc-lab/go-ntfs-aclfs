package go_ntfs_aclfs

import (
	"fmt"
	winsddlconverter "github.com/jc-lab/win-sddl-converter"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

type fsImpl struct {
	root    string
	options Options
}

func OpenFS(root string, options *Options) (FS, error) {
	rootPath, err := filepath.Abs(root)
	if err != nil {
		rootPath = root
	}

	//var fsType string
	//
	//mounts, err := mountinfo.GetMounts(func(info *mountinfo.Info) (skip, stop bool) {
	//	return !strings.HasPrefix(rootPath, info.Mountpoint), false
	//})
	//if err == nil {
	//	slices.SortFunc(mounts, func(a, b *mountinfo.Info) int {
	//		if len(a.Mountpoint) > len(b.Mountpoint) {
	//			return -1
	//		} else if len(a.Mountpoint) < len(b.Mountpoint) {
	//			return 1
	//		} else {
	//			return 0
	//		}
	//	})
	//	if len(mounts) > 0 {
	//		fsType = mounts[0].FSType
	//	}
	//}
	//
	//log.Println(fsType)
	// "fuseblk"

	impl := &fsImpl{
		root: rootPath,
	}
	if options != nil {
		impl.options = *options
	}
	return impl, nil
}

func (impl *fsImpl) Open(name string) (File, error) {
	return os.Open(filepath.Join(impl.root, name))
}

func (impl *fsImpl) OpenFile(name string, flag int, perm fs.FileMode) (File, error) {
	f, err := os.OpenFile(filepath.Join(impl.root, name), flag, perm)
	if err == nil {
		err = impl.Chmod(name, perm)
		if err != nil {
			f.Close()
		}
	}
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (impl *fsImpl) ReadDir(name string) ([]fs.DirEntry, error) {
	return os.ReadDir(filepath.Join(impl.root, name))
}

func (impl *fsImpl) Mkdir(name string, perm fs.FileMode) error {
	err := os.Mkdir(filepath.Join(impl.root, name), perm)
	if err == nil {
		err = impl.Chmod(name, perm)
	}
	return err
}

func (impl *fsImpl) MkdirAll(name string, perm fs.FileMode) error {
	return mkdirAll(filepath.Join(impl.root, name), perm, func(path string) error {
		return impl.chmodImpl(path, perm)
	})
}

func (impl *fsImpl) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(filepath.Join(impl.root, name))
}

func (impl *fsImpl) Chmod(name string, mode fs.FileMode) error {
	return impl.chmodImpl(filepath.Join(impl.root, name), mode)
}

func (impl *fsImpl) chmodImpl(path string, mode fs.FileMode) error {
	var useInherit bool
	if impl.options.UseInheritanceInDirectory {
		stat, err := os.Stat(path)
		if err == nil {
			useInherit = stat.IsDir()
		}
	}
	sddl := PermToSddl(mode, impl.options.OwnerSids, impl.options.GroupSids, impl.options.OtherPermissionPolicy, useInherit)
	return ChSddl(path, sddl)
}

func (impl *fsImpl) Rename(oldpath, newpath string) error {
	return os.Rename(filepath.Join(impl.root, oldpath), filepath.Join(impl.root, newpath))
}

func (impl *fsImpl) Remove(name string) error {
	return os.Remove(filepath.Join(impl.root, name))
}

func (impl *fsImpl) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return os.Chtimes(filepath.Join(impl.root, name), atime, mtime)
}

func (impl *fsImpl) Symlink(oldname, newname string) error {
	return os.Symlink(filepath.Join(impl.root, oldname), filepath.Join(impl.root, newname))
}

func (impl *fsImpl) ChSddl(name string, sddl string) error {
	return ChSddl(filepath.Join(impl.root, name), sddl)
}

func PermToSddl(mode fs.FileMode, ownerSids []string, groupSids []string, otherPermissionPolicy OtherPermissionPolicy, useInherit bool) string {
	var sddl strings.Builder
	var aceFlags string

	if useInherit {
		aceFlags = "OICI"
	}

	var ownerSid string

	if len(ownerSids) > 0 {
		ownerSid = ownerSids[0]
		sddl.WriteString("O:" + ownerSids[0])
	}
	if len(groupSids) > 0 {
		sddl.WriteString("G:" + groupSids[0])
	}
	sddl.WriteString("D:PAI")

	if len(ownerSids) > 0 {
		for _, sid := range ownerSids {
			sddl.WriteString("(A;" + aceFlags + ";" + fmt.Sprintf("0x%x", permToAccessMask(mode>>6)) + ";;;" + sid + ")")
		}
	} else {
		sddl.WriteString("(A;" + aceFlags + ";" + fmt.Sprintf("0x%x", permToAccessMask(mode>>6)) + ";;;CO)")
	}

	if len(groupSids) > 0 {
		for _, sid := range groupSids {
			if sid != ownerSid {
				sddl.WriteString("(A;" + aceFlags + ";" + fmt.Sprintf("0x%x", permToAccessMask(mode>>3)) + ";;;" + sid + ")")
			}
		}
	} else {
		sddl.WriteString("(A;" + aceFlags + ";" + fmt.Sprintf("0x%x", permToAccessMask(mode>>3)) + ";;;CG)")
	}

	for i := 0; i < 2; i++ {
		if (i == 0) && ((otherPermissionPolicy & OtherPermissionToEveryone) != 0) {
			sddl.WriteString("(A;" + aceFlags + ";" + fmt.Sprintf("0x%x", permToAccessMask(mode&0x7)) + ";;;WD)")
		}
		if (i == 1) && ((otherPermissionPolicy & OtherPermissionToUsers) != 0) {
			sddl.WriteString("(A;" + aceFlags + ";" + fmt.Sprintf("0x%x", permToAccessMask(mode&0x7)) + ";;;BU)")
		}
	}

	return sddl.String()
}

func permToAccessMask(mode fs.FileMode) uint32 {
	var accessMask uint32
	if (mode & 1) != 0 { // execute
		accessMask |= winsddlconverter.FILE_EXECUTE | winsddlconverter.FILE_READ_DATA
	}
	if (mode & 2) != 0 { // write
		accessMask |= winsddlconverter.FILE_WRITE_DATA | winsddlconverter.FILE_APPEND_DATA | winsddlconverter.FILE_WRITE_ATTRIBUTES | winsddlconverter.FILE_WRITE_EA | winsddlconverter.FILE_DELETE_CHILD
	}
	if (mode & 4) != 0 { // read
		accessMask |= winsddlconverter.FILE_READ_ACCESS
	}
	if (mode & 7) == 7 {
		accessMask |= winsddlconverter.FILE_ALL_ACCESS
	}

	// read security
	accessMask |= winsddlconverter.FILE_READ_EA

	return accessMask
}

func mkdirAll(path string, perm fs.FileMode, callback func(path string) error) error {
	// Fast path: if we can tell whether path is a directory or file, stop with success or error.
	dir, err := os.Stat(path)
	if err == nil {
		if dir.IsDir() {
			return nil
		}
		return &os.PathError{Op: "mkdir", Path: path, Err: syscall.ENOTDIR}
	}

	err = mkdirAll(filepath.Dir(path), perm, callback)
	if err != nil {
		return err
	}

	err = os.Mkdir(path, perm)
	if err == nil {
		err = callback(path)
	} else if os.IsExist(err) {
		err = nil
	}
	return err
}
