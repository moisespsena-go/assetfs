package assetfs

import "github.com/moisespsena/go-path-helpers"

var (
	pkg                 = path_helpers.GetCalledDir()
	CB_REPO_CONFIG_INIT = pkg + ":assets_init"
	CB_REPO_SYNC        = pkg + ":assets_sync"
)
