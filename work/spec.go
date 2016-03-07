package work

import (
	"github.com/blablacar/dgr/bin-dgr/common"
)

const PATH_ATTRIBUTES = "/attributes"

const ACTIVE_ACTIVE = "active"
const SUB_RUNNING = "running"
const PATH_SERVICE_MANIFEST = "/service-manifest.yml"
const NODE_HOSTNAME = "hostname"
const PATH_UNIT_SERVICE_TEMPLATE = "/unit.tmpl"
const PATH_UNIT_TIMER_TEMPLATE = "/unit.timer.tmpl"
const EARLY = true
const LATE = false

type Action int

const (
	ACTION_YES Action = iota
	ACTION_SKIP
	ACTION_DIFF
	ACTION_QUIT
)

type UnitStatus struct {
	Unit    string
	Machine string
	Active  string
	Sub     string
}

type HookInfo struct {
	Command    string
	Service    *Service
	Unit       *Unit
	Action     string
	Attributes string
}

type ServiceManifest struct {
	ConcurrentUpdater int                 `yaml:"concurrentUpdater"`
	Containers        []common.ACFullname `yaml:"containers"`
	ExecStartPre      []string            `yaml:"execStartPre"`
	ExecStart         []string            `yaml:"execStart"`
	Nodes             interface{}         `yaml:"nodes"`
}

type UnitType int

const (
	TYPE_SERVICE UnitType = iota
	TYPE_TIMER
)

func (u UnitType) String() string {
	switch u {
	case TYPE_SERVICE:
		return ".service"
	case TYPE_TIMER:
		return ".timer"
	}
	return ""
}
