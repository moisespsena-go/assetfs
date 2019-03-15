package assetfsapi

const (
	WalkDirs WalkMode = 1 << iota
	WalkFiles
	WalkNameSpaces
	WalkNameSpacesLookUp
	WalkParentLookUp
	WalkReverse

	WalkAll = WalkFiles | WalkDirs | WalkNameSpaces | WalkNameSpacesLookUp | WalkParentLookUp
)

type WalkMode int

func (f WalkMode) IsDirs() bool {
	return (f & WalkDirs) != 0
}

func (f WalkMode) IsFiles() bool {
	return (f & WalkFiles) != 0
}

func (f WalkMode) IsNameSpaces() bool {
	return (f & WalkNameSpaces) != 0
}

func (f WalkMode) IsNameSpacesLookUp() bool {
	return (f & WalkNameSpacesLookUp) != 0
}

func (f WalkMode) IsParentLookUp() bool {
	return (f & WalkParentLookUp) != 0
}
func (f WalkMode) IsReverse() bool {
	return (f & WalkReverse) != 0
}
