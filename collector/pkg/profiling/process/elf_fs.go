package process

import (
	"io"
	"io/fs"
	"path"
	"time"
)

// Open exposes the target's root to the fork's ELF/debuglink reader. Keep the
// existing openat2 root confinement; os.DirFS would allow symlink escapes.
func (sp *systemProcess) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}
	return openInProcRoot(sp.pid, "/"+name)
}

// Open exposes files reconstructed from a core through the same fs.FS contract.
// Each view has its own read offset and borrows the core's lifetime.
func (cd *CoredumpProcess) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}
	cf, ok := cd.files["/"+name]
	if !ok {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	var size uint64
	for _, m := range cf.Mappings {
		size = max(size, m.FileOffset+m.Prog.Filesz)
	}
	return &coreELFFile{
		SectionReader: io.NewSectionReader(cf, 0, int64(size)),
		file:          cf,
	}, nil
}

type coreELFFile struct {
	*io.SectionReader
	file *CoredumpFile
}

func (*coreELFFile) Close() error { return nil }

// ELFLoadHints preserves relocation and musl handling when the ELF is available
// only as loaded segments rather than as the original file on disk.
func (f *coreELFFile) ELFLoadHints() (uint64, bool) {
	return f.file.Base, f.file.parent.hasMusl
}

func (f *coreELFFile) Stat() (fs.FileInfo, error) {
	return coreELFFileInfo{name: path.Base(f.file.Name.String()), size: f.Size()}, nil
}

type coreELFFileInfo struct {
	name string
	size int64
}

func (i coreELFFileInfo) Name() string     { return i.name }
func (i coreELFFileInfo) Size() int64      { return i.size }
func (coreELFFileInfo) Mode() fs.FileMode  { return 0444 }
func (coreELFFileInfo) ModTime() time.Time { return time.Time{} }
func (coreELFFileInfo) IsDir() bool        { return false }
func (coreELFFileInfo) Sys() any           { return nil }
