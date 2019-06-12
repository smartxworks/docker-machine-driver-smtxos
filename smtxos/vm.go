package smtxos

import (
	"fmt"
)

const (
	VMStatusRunning   = "running"
	VMStatusStopped   = "stopped"
	VMStatusSuspended = "suspended"
)

type VM struct {
	UUID         string       `json:"uuid"`
	VMName       string       `json:"vm_name"`
	VCPU         int32        `json:"vcpu"`
	CPU          *VMCPU       `json:"cpu"`
	Memory       int64        `json:"memory"`
	Disks        []*VMDisk    `json:"disks"`
	NICs         []*VMNIC     `json:"nics"`
	AutoSchedule bool         `json:"auto_schedule"`
	HA           bool         `json:"ha"`
	Status       string       `json:"status,omitempty"`
	GuestInfo    *VMGuestInfo `json:"guest_info"`
}

type VMCPU struct {
	Topology *VMCPUTopology `json:"topology"`
}

type VMCPUTopology struct {
	Cores   int32 `json:"cores"`
	Sockets int32 `json:"sockets"`
}

type VMDisk struct {
	Type              string `json:"type"`
	Bus               string `json:"bus"`
	Name              string `json:"name"`
	SizeInByte        int64  `json:"size_in_byte,omitempty"`
	SrcExportID       string `json:"src_export_id,omitempty"`
	SrcInodeID        string `json:"src_inode_id,omitempty"`
	CloneBeforeCreate bool   `json:"clone_before_create,omitempty"`
	NewSizeInByte     int64  `json:"new_size_in_byte,omitempty"`
	StoragePolicyUUID string `json:"storage_policy_uuid,omitempty"`
}

type VMNIC struct {
	OVS      string       `json:"ovs"`
	VLANUUID string       `json:"vlan_uuid"`
	VLANs    []*VMNICVLAN `json:"vlans"`
}

type VMNICVLAN struct {
	VLANID int32 `json:"vlan_id"`
}

type VMGuestInfo struct {
	NICs []*VMGuestInfoNIC `json:"nics"`
}

type VMGuestInfoNIC struct {
	IPAddresses []VMGuestInfoNICIPAddress `json:"ip_addresses"`
}

type VMGuestInfoNICIPAddress struct {
	IPAddressType string `json:"ip_address_type"`
	IPAddress     string `json:"ip_address"`
}

func (c *client) GetVM(uuid string) (*VM, error) {
	var vm *VM
	if err := c.do("GET", fmt.Sprintf("/v2/vms/%s", uuid), nil, nil, &vm, true); err != nil {
		return nil, err
	}
	return vm, nil
}

func (c *client) CreateVM(vm *VM) (*Job, error) {
	var job *Job
	if err := c.do("POST", "/v2/vms", nil, vm, &job, true); err != nil {
		return nil, err
	}
	return job, nil
}

func (c *client) StartVM(uuid string) (*Job, error) {
	body := struct {
		AutoSchedule bool `json:"auto_schedule"`
	}{
		AutoSchedule: true,
	}
	var job *Job
	if err := c.do("POST", fmt.Sprintf("/v2/vms/%s/start", uuid), nil, &body, &job, true); err != nil {
		return nil, err
	}
	return job, nil
}

func (c *client) StopVM(uuid string) (*Job, error) {
	body := struct {
		Force bool `json:"force"`
	}{
		Force: true,
	}
	var job *Job
	if err := c.do("POST", fmt.Sprintf("/v2/vms/%s/stop", uuid), nil, &body, &job, true); err != nil {
		return nil, err
	}
	return job, nil
}

func (c *client) RebootVM(uuid string) (*Job, error) {
	body := struct {
		Force bool `json:"force"`
	}{
		Force: true,
	}
	var job *Job
	if err := c.do("POST", fmt.Sprintf("/v2/vms/%s/reboot", uuid), nil, &body, &job, true); err != nil {
		return nil, err
	}
	return job, nil
}

func (c *client) DeleteVM(uuid string) (*Job, error) {
	body := struct {
		IncludeVolumes bool `json:"include_volumes"`
	}{
		IncludeVolumes: true,
	}
	var job *Job
	if err := c.do("DELETE", fmt.Sprintf("/v2/vms/%s", uuid), nil, body, &job, true); err != nil {
		return nil, err
	}
	return job, nil
}

func (c *client) SetVMSSHPublicKey(vmUUID string, sshPublicKey string) error {
	body := struct {
		VMUUID       string `json:"vm_uuid"`
		SSHPublicKey string `json:"ssh_public_key"`
	}{
		VMUUID:       vmUUID,
		SSHPublicKey: sshPublicKey,
	}
	if err := c.do("POST", "/v2/vm_additional_info/set_ssh_public_key", nil, body, nil, true); err != nil {
		return err
	}
	return nil
}
