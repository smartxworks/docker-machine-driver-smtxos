package smtxos

import (
	"fmt"
	"net/http"
	"net/url"
)

type StoragePolicy struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
}

func (c *client) ListStoragePolicies() ([]*StoragePolicy, error) {
	var policies []*StoragePolicy
	if err := c.do(http.MethodGet, "/v2/storage_policies", nil, nil, &policies, true); err != nil {
		return nil, err
	}
	return policies, nil
}

type NFSExport struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (c *client) ListNFSExports() ([]*NFSExport, error) {
	var exports []*NFSExport
	if err := c.do(http.MethodGet, "/v2/nfs/exports", nil, nil, &exports, true); err != nil {
		return nil, err
	}
	return exports, nil
}

type NFSInode struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	ParentID   string `json:"parent_id"`
	SharedSize int64  `json:"shared_size"`
	UniqueSize int64  `json:"unique_size"`
}

func (c *client) ListNFSInodes(exportID string, parentID string) ([]*NFSInode, error) {
	params := url.Values{}
	if parentID != "" {
		params.Add("parent_id", parentID)
	} else {
		params.Add("parent_id", exportID[0:18])
	}
	var data struct {
		Inodes []*NFSInode `json:"inodes"`
	}
	if err := c.do(http.MethodGet, fmt.Sprintf("/v2/nfs/exports/%s/inodes", exportID), params, nil, &data, true); err != nil {
		return nil, err
	}
	return data.Inodes, nil
}
