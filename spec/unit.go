package spec

const PATH_UNIT_TEMPLATE = "/unit.tmpl"

type Unit interface {
	GetName() string
	GetService() Service
}
