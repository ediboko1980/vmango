package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)

type VZInfo struct {
	ID        int            `json:"ctid"`
	Name      string         `json:"hostname"`
	Status    string         `json:"status"`
	Cpus      int            `json:"cpus"`
	Physpages map[string]int `json:"physpages"`
}

type OVZStorage struct {
}

func NewOVZStorage() *OVZStorage {
	return &OVZStorage{}
}

func fillOvzVm(vm *VirtualMachine, info *VZInfo) {
	vm.Name = info.Name
	vm.Uuid = fmt.Sprintf("%d", info.ID)
	vm.Cpus = info.Cpus
	vm.Memory = info.Physpages["limit"]
	switch info.Status {
	default:
		vm.State = STATE_UNKNOWN
	case "running":
		vm.State = STATE_RUNNING
	}
}

func (store *OVZStorage) ListMachines(machines *[]*VirtualMachine) error {
	var out bytes.Buffer
	cmd := exec.Command("vzlist", "-j", "-a")
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return err
	}
	vzinfos := []*VZInfo{}
	if err := json.Unmarshal(out.Bytes(), &vzinfos); err != nil {
		return fmt.Errorf("failed to parse vzlist output: %s", err)
	}

	for _, vzinfo := range vzinfos {
		vm := &VirtualMachine{}
		fillOvzVm(vm, vzinfo)
		*machines = append(*machines, vm)
	}
	return nil

}

func (store *OVZStorage) GetMachine(machine *VirtualMachine) (bool, error) {
	if machine.Name == "" {
		return false, nil
	}

	var out bytes.Buffer
	cmd := exec.Command("vzlist", "-h", machine.Name, "-j", "-a")
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return false, err
	}
	vzinfos := []*VZInfo{}
	if err := json.Unmarshal(out.Bytes(), &vzinfos); err != nil {
		return false, fmt.Errorf("failed to parse vzlist output: %s", err)
	}

	if len(vzinfos) == 0 {
		return false, nil
	}

	if len(vzinfos) > 1 {
		return true, fmt.Errorf("%d containers found for name %s", len(vzinfos), machine.Name)
	}

	fillOvzVm(machine, vzinfos[0])
	return true, nil
}