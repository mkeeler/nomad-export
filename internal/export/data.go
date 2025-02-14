package export

import "github.com/hashicorp/nomad/api"

type Data struct {
	Namespaces map[string]NamespaceData `json:"namespaces,omitempty"`
}

type NamespaceData struct {
	Definition *api.Namespace `json:"definition,omitempty"`
	NamespacedData
}

type NamespacedData struct {
	Jobs map[string]JobData `json:"jobs,omitempty"`
}

type JobData struct {
	Info       *api.JobListStub `json:"info,omitempty"`
	Definition *api.Job         `json:"definition,omitempty"`
}
