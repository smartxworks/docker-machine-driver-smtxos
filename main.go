package main

import (
	"github.com/docker/machine/libmachine/drivers/plugin"
	"github.com/smartxworks/docker-machine-driver-smtxos/smtxos"
)

func main() {
	plugin.RegisterDriver(smtxos.NewDriver("", ""))
}
