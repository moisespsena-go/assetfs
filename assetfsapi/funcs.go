package assetfsapi

import "context"

type PathRegisterCallback = func(fs Interface)
type CbWalkFunc = func(name string, isDir bool) error
type CbWalkInfoFunc = func(info FileInfo) error
type AssetReaderFunc = func(name string) (data []byte, err error)
type AssetReaderFuncC = func(ctx context.Context, name string) (data []byte, err error)
type GetAssetInfoFunc = func(name string) (info FileInfo, err error)
type GetAssetInfoFuncC = func(ctx context.Context, name string) (info FileInfo, err error)
type PathFormatterFunc = func(pth *string)

type WalkFunc = func(pth string, cb CbWalkFunc, mode WalkMode) error
type WalkInfoFunc = func(path string, cb CbWalkInfoFunc, mode WalkMode) error
type GlobFunc = func(pattern GlobPattern, cb func(pth string, isDir bool) error) error
type GlobInfoFunc = func(pattern GlobPattern, cb func(info FileInfo) error) error
