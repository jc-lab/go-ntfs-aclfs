//go:build linux || unix
// +build linux unix

package go_ntfs_aclfs

import (
	winsddlconverter "github.com/jc-lab/win-sddl-converter"
	"golang.org/x/sys/unix"
)

func ChSddl(path string, sddl string) error {
	parsedSddl, err := winsddlconverter.ParseSDDL(sddl)
	if err != nil {
		return err
	}
	bin, err := parsedSddl.ToBinary()
	if err != nil {
		return err
	}
	return unix.Setxattr(path, "system.ntfs_acl", bin, 0)
}
