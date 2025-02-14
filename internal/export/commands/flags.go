package commands

import (
	"flag"
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/nomad/api"
)

type httpFlags struct {
	// client api flags
	address       stringValue
	token         stringValue
	tokenFile     stringValue
	caFile        stringValue
	caPath        stringValue
	certFile      stringValue
	keyFile       stringValue
	tlsServerName stringValue
}

func (f *httpFlags) flags() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.Var(&f.address, "address",
		"The address of the Nomad server.\n"+
			"Overrides the NOMAD_ADDR environment variable if set.\n"+
			"Default = http://127.0.0.1:4646")
	fs.Var(&f.token, "token",
		"The SecretID of an ACL token to use to authenticate API requests with.\n"+
			"Overrides the NOMAD_TOKEN environment variable if set.")
	fs.Var(&f.caFile, "ca-cert",
		"Path to a PEM encoded CA cert file to use to verify the Nomad server SSL certificate.\n"+
			"Overrides the NOMAD_CACERT environment variable if set.")
	fs.Var(&f.caPath, "ca-path",
		"Path to a directory of PEM encoded CA cert files to verify the Nomad server SSL certificate.\n"+
			"If both -ca-cert and -ca-path are specified, -ca-cert is used.\n"+
			"Overrides the NOMAD_CAPATH environment variable if set.")
	fs.Var(&f.certFile, "client-cert",
		"Path to a PEM encoded client certificate for TLS authentication to the Nomad server.\n"+
			"Must also specify -client-key.\n"+
			"Overrides the NOMAD_CLIENT_CERT environment variable if set.")
	fs.Var(&f.keyFile, "client-key",
		"Path to an unencrypted PEM encoded private key matching the client certificate form -client-cert.\n"+
			"Overrides the NOMAD_CLIENT_KEY environment variable if set.")
	fs.Var(&f.tlsServerName, "tls-server-name",
		"The server name to use as the SNI host when connecting via TLS.\n"+
			"Overrides the NOMAD_TLS_SERVER_NAME environment variable if set.")
	return fs
}

func (f *httpFlags) apiClient() (*api.Client, error) {
	c := api.DefaultConfig()

	f.mergeOntoConfig(c)

	return api.NewClient(c)
}

func (f *httpFlags) mergeOntoConfig(c *api.Config) {
	f.address.Merge(&c.Address)
	f.token.Merge(&c.SecretID)
	f.caFile.Merge(&c.TLSConfig.CACert)
	f.caPath.Merge(&c.TLSConfig.CAPath)
	f.certFile.Merge(&c.TLSConfig.ClientCert)
	f.keyFile.Merge(&c.TLSConfig.ClientKey)
	f.tlsServerName.Merge(&c.TLSConfig.TLSServerName)
}

func flagMerge(dst, src *flag.FlagSet) {
	if dst == nil {
		panic("dst cannot be nil")
	}
	if src == nil {
		return
	}
	src.VisitAll(func(f *flag.Flag) {
		dst.Var(f.Value, f.Name, f.Usage)
	})
}

// stringValue provides a flag value that's aware if it has been set.
type stringValue struct {
	v *string
}

// merge will overlay this value if it has been set.
func (s *stringValue) Merge(onto *string) {
	if s.v != nil {
		*onto = *(s.v)
	}
}

// Set implements the flag.Value interface.
func (s *stringValue) Set(v string) error {
	if s.v == nil {
		s.v = new(string)
	}
	*(s.v) = v
	return nil
}

// String implements the flag.Value interface.
func (s *stringValue) String() string {
	var current string
	if s.v != nil {
		current = *(s.v)
	}
	return current
}

type setValue struct {
	values  map[string]struct{}
	allowed map[string]struct{}
}

func (s *setValue) init(allowed []string) {
	s.values = make(map[string]struct{})
	s.allowed = make(map[string]struct{})
	for _, v := range allowed {
		s.allowed[v] = struct{}{}
	}
}

func (s *setValue) Set(v string) error {
	_, ok := s.allowed[v]
	if !ok {
		return fmt.Errorf("Value %q is not allowed. Allowed values are: %s", v, sliceString(s.allowed))
	}

	s.values[v] = struct{}{}
	return nil
}

func (s *setValue) String() string {
	return sliceString(s.values)
}

func sliceString(valueMap map[string]struct{}) string {
	var values []string
	for v := range valueMap {
		values = append(values, v)
	}

	slices.Sort(values)
	return strings.Join(values, ", ")
}
