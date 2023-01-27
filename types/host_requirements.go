package types

import (
	"fmt"
	"runtime"

	log "github.com/sirupsen/logrus"
	"github.com/srl-labs/containerlab/virt"
)

type HostRequirements struct {
	SSSE3 bool `json:"ssse3,omitempty"` // ssse3 cpu instruction
	// indicates that KVM virtualization is required for this node to run
	VirtRequired bool `json:"virt-required,omitempty"`
	// the minimum amount of vcpus this node requires
	MinVCPU           int           `json:"min-vcpu,omitempty"`
	MinVCPUFailAction FailBehaviour `json:"min-vcpu-fail-action,omitempty"`
	// The minimum amount of memory this node requires
	MinFreeMemoryGb           int           `json:"min-free-memory,omitempty"`
	MinFreeMemoryGbFailAction FailBehaviour `json:"min-free-memory-fail-action,omitempty"`
}

type FailBehaviour int

const (
	FailBehaviourLog FailBehaviour = iota
	FailBehaviourError
)

// NewHostRequirements is the constructor for new HostRequirements structs.
func NewHostRequirements() *HostRequirements {
	return &HostRequirements{
		MinVCPUFailAction:         FailBehaviourLog,
		MinFreeMemoryGbFailAction: FailBehaviourLog,
	}
}

func (h *HostRequirements) Verify() error {
	// check virtualization Support
	if h.VirtRequired && !virt.VerifyVirtSupport() {
		return fmt.Errorf("the CPU virtualization support is required, but not available")
	}
	// check SSSE3 support
	if h.SSSE3 && !virt.VerifySSSE3Support() {
		return fmt.Errorf("the SSSE3 CPU feature is required, but not available")
	}
	// check minimum vCPUs
	if valid, num := h.verifyMinVCpu(); !valid {
		message := fmt.Sprintf("the topology requires minimum %d vCPUs, but the host has %d vCPUs", h.MinVCPU, num)
		switch h.MinFreeMemoryGbFailAction {
		case FailBehaviourError:
			return fmt.Errorf(message)
		case FailBehaviourLog:
			log.Error(message)
		}
	}
	// check minimum FreeMemory
	if valid, num := h.verifyMinAvailMemory(); !valid {
		message := fmt.Sprintf("the defined minimum free memory based on the nodes in your topology is %d GB whilst only %d GB memory is free", h.MinFreeMemoryGb, num)
		switch h.MinFreeMemoryGbFailAction {
		case FailBehaviourError:
			return fmt.Errorf(message)
		case FailBehaviourLog:
			log.Error(message)
		}
	}
	return nil
}

// verifyMinAvailMemory verifies that the node requirement for minimum free memory is met.
// It returns a bool indicating if the requirement is met and the amount of available memory in GB.
func (h *HostRequirements) verifyMinAvailMemory() (bool, uint64) {
	freeMemG := virt.GetSysMemory(virt.MemoryTypeAvailable) / 1024 / 1024 / 1024

	// if the MinFreeMemory amount is 0, there is no requirement defined, so result is true
	if h.MinFreeMemoryGb == 0 {
		return true, freeMemG
	}

	// amount of Free Memory must be greater-equal the requirement
	boolResult := uint64(h.MinFreeMemoryGb) <= freeMemG
	return boolResult, freeMemG
}

// verifyMinVCpu verifies that the node requirement for minimum vCPU count is met.
func (h *HostRequirements) verifyMinVCpu() (bool, int) {
	numCpu := runtime.NumCPU()

	// if the minCPU amount is 0, there is no requirement defined, so result is true
	if h.MinVCPU == 0 {
		return true, numCpu
	}

	// count of vCPUs must be greater-equal the requirement
	boolResult := h.MinVCPU <= numCpu
	return boolResult, numCpu
}
