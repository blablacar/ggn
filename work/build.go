package work

type Flags struct {
	All                bool
	Yes                bool
	Force              bool
	GenerateManifests  []string
	ManifestAttributes string
}

var BuildFlags = Flags{}
