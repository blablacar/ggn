package work

type Flags struct {
	All               bool
	Yes               bool
	Force             bool
	GenerateManifests []string
}

var BuildFlags = Flags{}
