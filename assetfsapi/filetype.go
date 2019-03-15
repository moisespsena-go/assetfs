package assetfsapi

const (
	FileTypeNameSpace FileType = 1 << iota
	FileTypeNormal
	FileTypeDir
	FileTypeReal
	FileTypeBindata
)

type FileType int

func (f FileType) IsNameSpace() bool {
	return (f & FileTypeNameSpace) != 0
}

func (f FileType) IsNormal() bool {
	return (f & FileTypeNormal) != 0
}

func (f FileType) IsDir() bool {
	return (f & FileTypeDir) != 0
}

func (f FileType) IsReal() bool {
	return (f & FileTypeReal) != 0
}

func (f FileType) IsBindata() bool {
	return (f & FileTypeBindata) != 0
}
