package spec

const PATH_UNIT_SERVICE_TEMPLATE = "/unit.tmpl"
const PATH_UNIT_TIMER_TEMPLATE = "/unit.timer.tmpl"

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

type Unit interface {
	GetType() UnitType
	GetName() string
	GetService() Service
}
