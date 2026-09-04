package scripts

import "embed"

//go:embed *
var scripts embed.FS

func GetFs() embed.FS {
	return scripts
}
