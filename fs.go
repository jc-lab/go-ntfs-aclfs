package go_ntfs_aclfs

import (
	"io"
	"io/fs"
	"path/filepath"
)

//const (
//	// UsersReadExecuteDacl : read-only and executable for Users
//	// FR : SYNCHRONIZE + READ_CONTROL + FILE_READ_ATTRIBUTES + FILE_EXECUTE + FILE_READ_EA + FILE_READ_DATA
//	UsersReadExecuteDacl string = "(A;;FR;;;BU)"
//
//	// UsersReadDacl : read-only for Users
//	// 0x1200a9 : SYNCHRONIZE + READ_CONTROL + FILE_READ_ATTRIBUTES + FILE_EXECUTE + FILE_READ_EA + FILE_READ_DATA
//	UsersReadDacl string = "(A;;0x1200a9;;;BU)"
//)

type File interface {
	fs.File
	io.ReaderAt
	io.Writer
	io.WriterAt
}

type FS interface {
	Open(name string) (File, error)
	OpenFile(name string, flag int, perm fs.FileMode) (File, error)
	ReadDir(name string) ([]fs.DirEntry, error)
	Mkdir(name string, perm fs.FileMode) error
	MkdirAll(name string, perm fs.FileMode) error
	Stat(name string) (fs.FileInfo, error)
	Chmod(name string, mode fs.FileMode) error
	ChSddl(name string, sddl string) error
}

type OtherPermissionPolicy int

const (
	OtherPermissionToNothing  OtherPermissionPolicy = 0x00
	OtherPermissionToEveryone OtherPermissionPolicy = 0x01
	OtherPermissionToUsers    OtherPermissionPolicy = 0x02
)

type Options struct {
	// Set OwnerSid when creating a file or directory
	OwnerSid string
	// Set GroupSid when creating a file or directory
	GroupSid string

	OtherPermissionPolicy     OtherPermissionPolicy
	UseInheritanceInDirectory bool
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
