package go_ntfs_aclfs

import (
	"io"
	"io/fs"
	"time"
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
	Name() string
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
	Rename(oldpath, newpath string) error
	Remove(name string) error
	Chtimes(name string, atime time.Time, mtime time.Time) error
	Symlink(oldname, newname string) error
	ChSddl(name string, sddl string) error
}

type OtherPermissionPolicy int

const (
	OtherPermissionToNothing  OtherPermissionPolicy = 0x00
	OtherPermissionToEveryone OtherPermissionPolicy = 0x01
	OtherPermissionToUsers    OtherPermissionPolicy = 0x02
)

type Options struct {
	// Set OwnerSids[0] when creating a file or directory
	OwnerSids []string
	// Set GroupSids[0] when creating a file or directory
	GroupSids []string

	OtherPermissionPolicy     OtherPermissionPolicy
	UseInheritanceInDirectory bool
}
