package smtxos

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

type VDS struct {
	UUID      string `json:"uuid"`
	Name      string `json:"name"`
	OVSBRName string `json:"ovsbr_name"`
}

func (c *client) ListVDSs() ([]*VDS, error) {
	var vdss []*VDS
	if err := c.do("GET", "/v2/network/vds", nil, nil, &vdss, true); err != nil {
		return nil, err
	}
	return vdss, nil
}

type VLAN struct {
	VDSUUID string `json:"vds_uuid"`
	UUID    string `json:"uuid"`
	Name    string `json:"name"`
	VLANID  int32  `json:"vlan_id"`
}

func (vlan *VLAN) UnmarshalJSON(data []byte) error {
	type originVLAN VLAN
	v := struct {
		*originVLAN
		VLANID interface{} `json:"vlan_id"`
	}{
		originVLAN: (*originVLAN)(vlan),
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	switch vlanID := v.VLANID.(type) {
	case string:
		i, err := strconv.Atoi(vlanID)
		if err != nil {
			return err
		}
		vlan.VLANID = int32(i)
	case int:
		vlan.VLANID = int32(vlanID)
	case int32:
		vlan.VLANID = int32(vlanID)
	case int64:
		vlan.VLANID = int32(vlanID)
	case float32:
		vlan.VLANID = int32(vlanID)
	case float64:
		vlan.VLANID = int32(vlanID)
	default:
		return errors.New("unsupported type of vlan_id")
	}
	return nil
}

func (c *client) ListVLANs(vdsUUID string) ([]*VLAN, error) {
	var data *struct {
		Entities []*VLAN `json:"entities"`
	}
	if err := c.do("GET", fmt.Sprintf("/v2/network/vds/%s/vlans", vdsUUID), nil, nil, &data, true); err != nil {
		return nil, err
	}
	return data.Entities, nil
}
