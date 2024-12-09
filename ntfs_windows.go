package go_ntfs_aclfs

import (
	"golang.org/x/sys/windows"
)

func ChSddl(path string, sddl string) error {
	securityDescriptor, err := windows.SecurityDescriptorFromString(sddl)
	if err != nil {
		return err
	}

	return SetSecurityDescriptor(path, securityDescriptor)
}

func SetSecurityDescriptor(filePath string, sd *windows.SECURITY_DESCRIPTOR) error {
	owner, _, err := sd.Owner()
	if err != nil {
		//Do not set partial values.
		owner = nil
	}
	group, _, err := sd.Group()
	if err != nil {
		//Do not set partial values.
		group = nil
	}
	dacl, _, err := sd.DACL()
	if err != nil {
		//Do not set partial values.
		dacl = nil
	}
	sacl, _, err := sd.SACL()
	if err != nil {
		//Do not set partial values.
		sacl = nil
	}

	return windows.SetNamedSecurityInfo(
		filePath,
		windows.SE_FILE_OBJECT,
		windows.OWNER_SECURITY_INFORMATION|windows.GROUP_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION|windows.SACL_SECURITY_INFORMATION|windows.LABEL_SECURITY_INFORMATION|windows.ATTRIBUTE_SECURITY_INFORMATION|windows.SCOPE_SECURITY_INFORMATION|windows.BACKUP_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION|windows.PROTECTED_SACL_SECURITY_INFORMATION|windows.UNPROTECTED_DACL_SECURITY_INFORMATION|windows.UNPROTECTED_SACL_SECURITY_INFORMATION,
		owner,
		group,
		dacl,
		sacl,
	)
}
