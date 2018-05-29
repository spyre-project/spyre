package platform

import (
	"github.com/spf13/afero"

	"github.com/dcso/spyre/sys"

	"os"
	"syscall"
)

func SkipDir(fs afero.Fs, path string) bool {
	file, err := fs.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()
	f, ok := file.(*os.File)
	if !ok {
		return false
	}
	var buf syscall.Statfs_t
	if err := syscall.Fstatfs(int(f.Fd()), &buf); err != nil {
		return false
	}
	switch uint32(buf.Type) {
	case
		// pseudo filesystems
		sys.BDEVFS_MAGIC,
		sys.BINFMTFS_MAGIC,
		sys.CGROUP_SUPER_MAGIC,
		sys.DEBUGFS_MAGIC,
		sys.EFIVARFS_MAGIC,
		sys.FUTEXFS_SUPER_MAGIC,
		sys.HUGETLBFS_MAGIC,
		sys.PIPEFS_MAGIC,
		sys.PROC_SUPER_MAGIC,
		sys.SELINUX_MAGIC,
		sys.SMACK_MAGIC,
		sys.SYSFS_MAGIC,
		// network filesystems
		sys.AFS_FS_MAGIC,
		sys.OPENAFS_FS_MAGIC,
		sys.CEPH_SUPER_MAGIC,
		sys.CIFS_MAGIC_NUMBER,
		sys.CODA_SUPER_MAGIC,
		sys.NCP_SUPER_MAGIC,
		sys.NFS_SUPER_MAGIC,
		sys.OCFS2_SUPER_MAGIC,
		sys.SMB_SUPER_MAGIC,
		sys.V9FS_MAGIC,
		sys.VMBLOCK_SUPER_MAGIC,
		sys.XENFS_SUPER_MAGIC:
		return true
	}
	return false
}
