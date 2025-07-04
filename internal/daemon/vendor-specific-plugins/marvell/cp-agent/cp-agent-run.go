package main

import (
	mrvlutils "github.com/openshift/dpu-operator/internal/daemon/vendor-specific-plugins/marvell/mrvl-utils"
	"k8s.io/klog/v2"
	"os/exec"
)

func setupDpuLink() error {
	err := mrvlutils.SetupHugepages()
	if err != nil {
		klog.Errorf("Failed to set up hugepages: %v", err)
		return err
	}

	cmd := "chroot /host modprobe vfio-pci"
	err = exec.Command("bash", "-c", cmd).Run()
	if err != nil {
		klog.Errorf("Failed to load driver vfio-pci: %v", err)
		return err
	}

	dpi_pf, err := mrvlutils.GetAllVfsByDeviceID(mrvlutils.MrvlDPIPFId)
	if err != nil {
		klog.Errorf("DPI PF not found: %v", err)
		return err
	}

	klog.Infof("Found DPI PF: %v", dpi_pf[0])

	err = mrvlutils.BindToVFIO(dpi_pf[0])
	if err != nil {
		klog.Errorf("Failed to bind DPI PF with VFIO: %v", err)
		return err
	}

	pem_pf, err := mrvlutils.GetAllVfsByDeviceID(mrvlutils.MrvlPEMPFId)
	if err != nil {
		klog.Errorf("PEM PF not found: %v", err)
		return err
	}

	klog.Infof("Found PEM PF: %v", pem_pf)

	err = mrvlutils.BindToVFIO(pem_pf[0])
	if err != nil {
		klog.Errorf("Failed to bind PEM PF with VFIO: %v", err)
		return err
	}

	cpagentCmd := "/usr/bin/octep_cp_agent /usr/bin/cn106xx.cfg --"
	cpagentCmd += " --dpi_dev " + dpi_pf[0]
	cpagentCmd += " --pem_dev " + pem_pf[0]
	cpagentCmd += " &> /tmp/octep-cp-agent.log"
	err = exec.Command("bash", "-c", cpagentCmd).Run()
	if err != nil {
		klog.Errorf("Failed to start cp-agent: %v", err)
		return err
	}

	return nil
}

func main() {
	if mrvlutils.DetectPlatformMode() == "dpu" {
		err := setupDpuLink()
		if err != nil {
			klog.Errorf("Failed to set up Control Plane Agent: %v", err)
		}
	} else {
		klog.Error("DPU support not enabled")
	}
}
