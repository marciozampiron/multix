package plugin

type Manifest struct {
	Name        string
	Version     string
	Author      string
	Description string
	EntryPoint  string
}

type Plugin struct {
	Manifest Manifest
	State    string
}
