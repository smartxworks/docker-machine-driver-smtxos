package smtxos

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"
)

type Driver struct {
	*drivers.BaseDriver

	Server            string
	Port              int32
	Username          string
	Password          string
	CPUCount          int32
	MemorySizeBytes   int64
	DiskSizeBytes     int64
	StoragePolicyName string
	DockerOSImagePath string
	NetworkName       string
	HA                bool

	UUID string

	client Client
}

func NewDriver(hostName string, storePath string) *Driver {
	return &Driver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: hostName,
			StorePath:   storePath,
		},
	}
}

func (d *Driver) DriverName() string {
	return "smtxos"
}

func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			EnvVar: "SMTXOS_SERVER",
			Name:   "smtxos-server",
			Usage:  "address of SMTX OS server",
			Value:  "",
		},
		mcnflag.IntFlag{
			EnvVar: "SMTXOS_PORT",
			Name:   "smtxos-port",
			Usage:  "port of SMTX OS server",
			Value:  80,
		},
		mcnflag.StringFlag{
			EnvVar: "SMTXOS_USERNAME",
			Name:   "smtxos-username",
			Usage:  "username used to login SMTX OS",
			Value:  "root",
		},
		mcnflag.StringFlag{
			EnvVar: "SMTXOS_PASSWORD",
			Name:   "smtxos-password",
			Usage:  "password used to login SMTX OS",
			Value:  "",
		},
		mcnflag.IntFlag{
			EnvVar: "SMTXOS_CPU_COUNT",
			Name:   "smtxos-cpu-count",
			Usage:  "number of CPU cores for VM",
			Value:  2,
		},
		mcnflag.IntFlag{
			EnvVar: "SMTXOS_MEMORY_SIZE",
			Name:   "smtxos-memory-size",
			Usage:  "size of memory for VM (in MB)",
			Value:  4096,
		},
		mcnflag.IntFlag{
			EnvVar: "SMTXOS_DISK_SIZE",
			Name:   "smtxos-disk-size",
			Usage:  "size of disk for VM (in MB)",
			Value:  10240,
		},
		mcnflag.StringFlag{
			EnvVar: "SMTXOS_STORAGE_POLICY_NAME",
			Name:   "smtxos-storage-policy-name",
			Usage:  "name of storage policy of disk for VM",
			Value:  "default",
		},
		mcnflag.StringFlag{
			EnvVar: "SMTXOS_DOCKEROS_IMAGE_PATH",
			Name:   "smtxos-dockeros-image-path",
			Usage:  "path of DockerOS image on SMTX OS, in the format of [datastore-name]/image-file-path",
			Value:  "[kubernetes]/SMTX-DockerOS.raw",
		},
		mcnflag.StringFlag{
			EnvVar: "SMTXOS_NETWORK_NAME",
			Name:   "smtxos-network-name",
			Usage:  "network name for VM",
			Value:  "default",
		},
		mcnflag.BoolFlag{
			EnvVar: "SMTXOS_HA",
			Name:   "smtxos-ha",
			Usage:  "whether to enable high availability for VM",
		},
	}
}

func (d *Driver) SetConfigFromFlags(opts drivers.DriverOptions) error {
	d.Server = opts.String("smtxos-server")
	d.Port = int32(opts.Int("smtxos-port"))
	d.Username = opts.String("smtxos-username")
	d.Password = opts.String("smtxos-password")
	d.CPUCount = int32(opts.Int("smtxos-cpu-count"))
	d.MemorySizeBytes = int64(opts.Int("smtxos-memory-size")) * 1024 * 1024
	d.DiskSizeBytes = int64(opts.Int("smtxos-disk-size")) * 1024 * 1024
	d.StoragePolicyName = opts.String("smtxos-storage-policy-name")
	d.DockerOSImagePath = opts.String("smtxos-dockeros-image-path")
	d.NetworkName = opts.String("smtxos-network-name")
	d.HA = opts.Bool("smtxos-ha")
	d.SSHUser = "centos"
	return nil
}

func (d *Driver) Create() error {
	if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
		return err
	}

	sshPublicKey, err := ioutil.ReadFile(d.GetSSHKeyPath() + ".pub")
	if err != nil {
		return err
	}

	vm := &VM{
		VMName: d.MachineName,
		VCPU:   d.CPUCount,
		CPU: &VMCPU{
			Topology: &VMCPUTopology{
				Cores:   d.CPUCount,
				Sockets: 1,
			},
		},
		Memory:       d.MemorySizeBytes,
		AutoSchedule: true,
		HA:           d.HA,
		Status:       VMStatusStopped,
	}

	exports, err := d.getClient().ListNFSExports()
	if err != nil {
		return err
	}

	osDisk := VMDisk{
		Type:              "disk",
		Bus:               "virtio",
		Name:              fmt.Sprintf("%s-os", d.MachineName),
		CloneBeforeCreate: true,
	}

	for _, export := range exports {
		if strings.HasPrefix(d.DockerOSImagePath, fmt.Sprintf("[%s]", export.Name)) {
			osDisk.SrcExportID = export.ID
			path := strings.TrimPrefix(d.DockerOSImagePath, fmt.Sprintf("[%s]", export.Name))
			inodeNames := strings.Split(path, "/")
			parentID := ""
			for _, inodeName := range inodeNames {
				inodeName = strings.TrimSpace(inodeName)
				if inodeName == "" {
					continue
				}

				inodes, err := d.getClient().ListNFSInodes(export.ID, parentID)
				if err != nil {
					return err
				}

				found := false
				for _, inode := range inodes {
					if inode.Name == inodeName {
						osDisk.SrcInodeID = inode.ID
						osDisk.NewSizeInByte = inode.SharedSize + inode.UniqueSize
						parentID = inode.ID
						found = true
						break
					}
				}

				if !found {
					return fmt.Errorf("DockerOS image '%s' not found", d.DockerOSImagePath)
				}
			}
			break
		}
	}

	vm.Disks = append(vm.Disks, &osDisk)

	dockerDisk := VMDisk{
		Type:       "disk",
		Bus:        "virtio",
		Name:       fmt.Sprintf("%s-docker", d.MachineName),
		SizeInByte: d.DiskSizeBytes,
	}

	policies, err := d.getClient().ListStoragePolicies()
	if err != nil {
		return err
	}

	found := false
	for _, policy := range policies {
		if policy.Name == d.StoragePolicyName {
			dockerDisk.StoragePolicyUUID = policy.UUID
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("storage policy '%s' not found", d.StoragePolicyName)
	}

	vm.Disks = append(vm.Disks, &dockerDisk)

	for _, networkName := range []string{d.NetworkName, "ovsbr-internal-default-network"} {
		vdss, err := d.getClient().ListVDSs()
		if err != nil {
			return err
		}

		found := false
		for _, vds := range vdss {
			vlans, err := d.getClient().ListVLANs(vds.UUID)
			if err != nil {
				return err
			}

			for _, vlan := range vlans {
				if vlan.Name == networkName {
					vm.NICs = append(vm.NICs, &VMNIC{
						OVS:      vds.OVSBRName,
						VLANUUID: vlan.UUID,
						VLANs: []*VMNICVLAN{
							&VMNICVLAN{
								VLANID: vlan.VLANID,
							},
						},
					})
					found = true
					break
				}
			}

			if found {
				break
			}
		}

		if !found {
			return fmt.Errorf("network '%s' not found", networkName)
		}
	}

	job, err := d.getClient().CreateVM(vm)
	if err != nil {
		return err
	}

	if err := d.waitJobDone(job.JobID); err != nil {
		return err
	}

	job, err = d.getClient().GetJob(job.JobID)
	if err != nil {
		return err
	}
	resources := job.Resources.(map[string]interface{})
	for _, v := range resources {
		resource := v.(map[string]interface{})
		if resource["type"] != "KVM_VM" {
			continue
		}
		d.UUID = resource["uuid"].(string)
	}

	if err := d.getClient().SetVMSSHPublicKey(d.UUID, string(sshPublicKey)); err != nil {
		return err
	}

	return d.Start()
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("tcp://%s:2375", ip), nil
}

func (d *Driver) GetIP() (string, error) {
	if d.IPAddress != "" {
		return d.IPAddress, nil

	}

	vm, err := d.getClient().GetVM(d.UUID)
	if err != nil {
		return "", err
	}

	if vm.GuestInfo != nil {
		for _, nic := range vm.GuestInfo.NICs {
			for _, ip := range nic.IPAddresses {
				if ip.IPAddressType != "ipv4" {
					continue
				}
				if ip.IPAddress == "127.0.0.1" || strings.HasPrefix(ip.IPAddress, "169.254.") {
					continue
				}

				d.IPAddress = ip.IPAddress
				return d.IPAddress, nil
			}
		}
	}

	return "", errors.New("IP address is not set")
}

func (d *Driver) GetState() (state.State, error) {
	vm, err := d.getClient().GetVM(d.UUID)
	if err != nil {
		return state.None, err
	}

	switch vm.Status {
	case VMStatusRunning:
		return state.Running, nil
	case VMStatusStopped:
		return state.Stopped, nil
	case VMStatusSuspended:
		return state.Paused, nil
	default:
		return state.Error, nil
	}
}

func (d *Driver) Start() error {
	job, err := d.getClient().StartVM(d.UUID)
	if err != nil {
		return err
	}
	return d.waitJobDone(job.JobID)
}

func (d *Driver) Stop() error {
	job, err := d.getClient().StopVM(d.UUID)
	if err != nil {
		return err
	}
	return d.waitJobDone(job.JobID)
}

func (d *Driver) Kill() error {
	return d.Stop()
}

func (d *Driver) Restart() error {
	job, err := d.getClient().RebootVM(d.UUID)
	if err != nil {
		return err
	}
	return d.waitJobDone(job.JobID)
}

func (d *Driver) Remove() error {
	job, err := d.getClient().DeleteVM(d.UUID)
	if err != nil {
		return err
	}
	return d.waitJobDone(job.JobID)
}

func (d *Driver) waitJobDone(id string) error {
	client := d.getClient()

	for {
		job, err := client.GetJob(id)
		if err != nil {
			return err
		}

		switch job.State {
		case JobStateDone:
			return nil
		case JobStateFailed:
			return errors.New(job.Description)
		}

		time.Sleep(time.Second)
	}
}

func (d *Driver) getClient() Client {
	if d.client == nil {
		d.client = NewClient(d.Server, d.Port, d.Username, d.Password)
	}
	return d.client
}
