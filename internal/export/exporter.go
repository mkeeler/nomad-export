package export

import (
	"fmt"

	"github.com/hashicorp/nomad/api"
)

func Export(client *api.Client, excludeTypes map[string]struct{}) (*Data, error) {
	e, err := newExporter(client, excludeTypes)
	if err != nil {
		return nil, err
	}
	return e.Export()
}

type exporter struct {
	client       *api.Client
	excludeTypes map[string]struct{}
}

func newExporter(client *api.Client, excludeTypes map[string]struct{}) (*exporter, error) {
	return &exporter{
		client:       client,
		excludeTypes: excludeTypes,
	}, nil
}

func (e *exporter) Export() (*Data, error) {
	ndata, err := e.exportNamespaces()
	if err != nil {
		return nil, fmt.Errorf("error exporting partitions: %w", err)
	}

	return &Data{
		Namespaces: ndata,
	}, nil
}

func (e *exporter) exportNamespaces() (map[string]NamespaceData, error) {
	namespaces, _, err := e.client.Namespaces().List(nil)
	if err != nil {
		return nil, fmt.Errorf("error listing namespaces: %w", err)
	}
	nmap := make(map[string]NamespaceData)

	for _, namespace := range namespaces {
		ndata, err := e.exportNamespace(namespace.Name)
		if err != nil {
			return nil, fmt.Errorf("error exporting namespace %s: %w", namespace.Name, err)
		}

		nmap[namespace.Name] = NamespaceData{
			Definition:     namespace,
			NamespacedData: *ndata,
		}
	}

	return nmap, nil
}

func (e *exporter) exportNamespace(namespace string) (*NamespacedData, error) {
	jobs, err := e.exportJobs(namespace)
	if err != nil {
		return nil, fmt.Errorf("error exporting jobs for namespace %s: %w", namespace, err)
	}

	return &NamespacedData{
		Jobs: jobs,
	}, nil
}

func (e *exporter) exportJobs(namespace string) (map[string]JobData, error) {
	opts := &api.QueryOptions{Namespace: namespace}
	jobs, _, err := e.client.Jobs().List(opts)
	if err != nil {
		return nil, fmt.Errorf("error listing jobs for namespace %s: %w", namespace, err)
	}

	jmap := make(map[string]JobData)

	for _, job := range jobs {
		definition, _, err := e.client.Jobs().Info(job.ID, opts)
		if err != nil {
			return nil, fmt.Errorf("error getting job %s (%s) definition in namespace %s: %w", job.Name, job.ID, namespace, err)
		}

		jmap[job.ID] = JobData{
			Info:       job,
			Definition: definition,
		}
	}

	return jmap, nil
}
