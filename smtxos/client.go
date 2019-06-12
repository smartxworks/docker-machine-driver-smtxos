package smtxos

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type Client interface {
	GetVM(uuid string) (*VM, error)
	CreateVM(vm *VM) (*Job, error)
	StartVM(uuid string) (*Job, error)
	StopVM(uuid string) (*Job, error)
	RebootVM(uuid string) (*Job, error)
	DeleteVM(uuid string) (*Job, error)
	SetVMSSHPublicKey(vmUUID string, sshPublicKey string) error
	ListVDSs() ([]*VDS, error)
	ListVLANs(vdsUUID string) ([]*VLAN, error)
	ListStoragePolicies() ([]*StoragePolicy, error)
	ListNFSExports() ([]*NFSExport, error)
	ListNFSInodes(exportID string, parentID string) ([]*NFSInode, error)
	GetJob(id string) (*Job, error)
}

type client struct {
	server     string
	port       int32
	username   string
	password   string
	token      string
	httpClient *http.Client
}

func NewClient(server string, port int32, username string, password string) Client {
	return &client{
		server:     server,
		port:       port,
		username:   username,
		password:   password,
		httpClient: &http.Client{},
	}
}

func (c *client) do(method string, path string, query url.Values, body interface{}, data interface{}, autoLogin bool) error {
	req := &http.Request{
		Method: method,
	}

	if c.token != "" {
		req.Header = make(http.Header)
		req.Header.Add("X-SmartX-Token", c.token)
		req.Header.Add("Grpc-Metadata-Token", c.token)
	}

	u, err := url.Parse(fmt.Sprintf("http://%s:%d/api%s", c.server, c.port, path))
	if err != nil {
		return err
	}

	if query != nil {
		u.RawQuery = query.Encode()
	}
	req.URL = u

	if body != nil {
		bs, err := json.Marshal(body)
		if err != nil {
			return err
		}
		req.Body = ioutil.NopCloser(bytes.NewBuffer(bs))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		if !autoLogin || resp.StatusCode != http.StatusUnauthorized {
			return errors.New(http.StatusText(resp.StatusCode))
		}

		if err := c.login(); err != nil {
			return err
		}
		return c.do(method, path, query, body, data, false)
	}

	if strings.HasPrefix(path, "/v2") {
		var result struct {
			EC    string          `json:"ec"`
			Data  json.RawMessage `json:"data"`
			Error interface{}     `json:"error"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return err
		}

		if result.EC != "EOK" {
			return fmt.Errorf("%v", result.Error)
		}

		if data == nil {
			return nil
		}

		if err := json.Unmarshal([]byte(result.Data), data); err != nil {
			return err
		}
	} else {
		if data == nil {
			return nil
		}

		if err := json.NewDecoder(resp.Body).Decode(data); err != nil {
			return err
		}
	}

	return nil
}

func (c *client) login() error {
	body := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		Username: c.username,
		Password: c.password,
	}

	var session struct {
		Token string `json:"token"`
	}

	if err := c.do("POST", "/v3/sessions", nil, body, &session, false); err != nil {
		return err
	}

	c.token = session.Token
	return nil
}
