package assetfs

type Traversable struct {
	AssetGetterInterface
	WalkFilesFunc     func(dir string, cb WalkFunc, onError ...OnErrorFunc) error
	WalkFilesInfoFunc func(dir string, cb WalkInfoFunc, onError ...OnErrorFunc) error
}

func (t *Traversable) WalkFiles(dir string, cb WalkFunc, onError ...OnErrorFunc) error {
	return t.WalkFilesFunc(dir, cb, onError...)
}

func (t *Traversable) WalkFilesInfo(dir string, cb WalkInfoFunc, onError ...OnErrorFunc) error {
	return t.WalkFilesInfoFunc(dir, cb, onError...)
}
