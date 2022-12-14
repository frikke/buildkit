package client

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/docker/docker/api"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/versions"
	"github.com/docker/go-connections/sockets"
	"github.com/pkg/errors"
)

const defaultDockerHost = "unix:///var/run/docker.sock"
const defaultProto = "unix"
const defaultAddr = "/var/run/docker.sock"

var ErrRedirect = errors.New("unexpected redirect in response")

type Client struct {
	scheme            string
	host              string
	proto             string
	addr              string
	basePath          string
	client            *http.Client
	version           string
	customHTTPHeaders map[string]string
	manualOverride    bool
	negotiateVersion  bool
	negotiated        bool
}

func CheckRedirect(req *http.Request, via []*http.Request) error {
	if via[0].Method == http.MethodGet {
		return http.ErrUseLastResponse
	}
	return ErrRedirect
}

func NewClientWithOpts(ops ...Opt) (*Client, error) {
	client, err := defaultHTTPClient(defaultDockerHost)
	if err != nil {
		return nil, err
	}
	c := &Client{
		host:    defaultDockerHost,
		version: api.DefaultVersion,
		client:  client,
		proto:   defaultProto,
		addr:    defaultAddr,
	}

	for _, op := range ops {
		if err := op(c); err != nil {
			return nil, err
		}
	}

	if c.scheme == "" {
		c.scheme = "http"

		tlsConfig := resolveTLSConfig(c.client.Transport)
		if tlsConfig != nil {
			// TODO(stevvooe): This isn't really the right way to write clients in Go.
			// `NewClient` should probably only take an `*http.Client` and work from there.
			// Unfortunately, the model of having a host-ish/url-thingy as the connection
			// string has us confusing protocol and transport layers. We continue doing
			// this to avoid breaking existing clients but this should be addressed.
			c.scheme = "https"
		}
	}

	return c, nil
}

func resolveTLSConfig(transport http.RoundTripper) *tls.Config {
	switch tr := transport.(type) {
	case *http.Transport:
		return tr.TLSClientConfig
	default:
		return nil
	}
}

func defaultHTTPClient(host string) (*http.Client, error) {
	hostURL, err := ParseHostURL(host)
	if err != nil {
		return nil, err
	}
	transport := &http.Transport{}
	_ = sockets.ConfigureTransport(transport, hostURL.Scheme, hostURL.Host)
	return &http.Client{
		Transport:     transport,
		CheckRedirect: CheckRedirect,
	}, nil
}

func (cli *Client) Close() error {
	if t, ok := cli.client.Transport.(*http.Transport); ok {
		t.CloseIdleConnections()
	}
	return nil
}

func (cli *Client) getAPIPath(ctx context.Context, p string, query url.Values) string {
	var apiPath string
	if cli.negotiateVersion && !cli.negotiated {
		cli.NegotiateAPIVersion(ctx)
	}
	if cli.version != "" {
		v := strings.TrimPrefix(cli.version, "v")
		apiPath = path.Join(cli.basePath, "/v"+v, p)
	} else {
		apiPath = path.Join(cli.basePath, p)
	}
	return (&url.URL{Path: apiPath, RawQuery: query.Encode()}).String()
}

func (cli *Client) ClientVersion() string {
	return cli.version
}

func (cli *Client) NegotiateAPIVersion(ctx context.Context) {
	if !cli.manualOverride {
		ping, _ := cli.Ping(ctx)
		cli.negotiateAPIVersionPing(ping)
	}
}

func (cli *Client) NegotiateAPIVersionPing(pingResponse types.Ping) {
	if !cli.manualOverride {
		cli.negotiateAPIVersionPing(pingResponse)
	}
}

func (cli *Client) negotiateAPIVersionPing(pingResponse types.Ping) {
	// default to the latest version before versioning headers existed
	if pingResponse.APIVersion == "" {
		pingResponse.APIVersion = "1.24"
	}

	// if the client is not initialized with a version, start with the latest supported version
	if cli.version == "" {
		cli.version = api.DefaultVersion
	}

	// if server version is lower than the client version, downgrade
	if versions.LessThan(pingResponse.APIVersion, cli.version) {
		cli.version = pingResponse.APIVersion
	}

	// Store the results, so that automatic API version negotiation (if enabled)
	// won't be performed on the next request.
	if cli.negotiateVersion {
		cli.negotiated = true
	}
}

func (cli *Client) DaemonHost() string {
	return cli.host
}

func (cli *Client) HTTPClient() *http.Client {
	c := *cli.client
	return &c
}

func ParseHostURL(host string) (*url.URL, error) {
	protoAddrParts := strings.SplitN(host, "://", 2)
	if len(protoAddrParts) == 1 {
		return nil, errors.Errorf("unable to parse docker host `%s`", host)
	}

	var basePath string
	proto, addr := protoAddrParts[0], protoAddrParts[1]
	if proto == "tcp" {
		parsed, err := url.Parse("tcp://" + addr)
		if err != nil {
			return nil, err
		}
		addr = parsed.Host
		basePath = parsed.Path
	}
	return &url.URL{
		Scheme: proto,
		Host:   addr,
		Path:   basePath,
	}, nil
}

func (cli *Client) Dialer() func(context.Context) (net.Conn, error) {
	return func(ctx context.Context) (net.Conn, error) {
		if transport, ok := cli.client.Transport.(*http.Transport); ok {
			if transport.DialContext != nil && transport.TLSClientConfig == nil {
				return transport.DialContext(ctx, cli.proto, cli.addr)
			}
		}
		return fallbackDial(cli.proto, cli.addr, resolveTLSConfig(cli.client.Transport))
	}
}
