package gitbackend

type FileInfo struct {
  name string // base name of the file
  // path  string
  // size  int64 // length in bytes for regular files; system-dependent for others
  isDir bool  // abbreviation for Mode().IsDir()
}

func (f *FileInfo) Name() string {
  return f.name
}

func (f *FileInfo) IsDir() bool {
  return f.isDir
}
