package platform

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/jaypipes/ghw"
)

type Platform interface {
	PciDevices() ([]*ghw.PCIDevice, error)
	NetDevs() ([]*ghw.NIC, error)
	Product() (*ghw.ProductInfo, error)
	ReadDeviceSerialNumber(pciDevice *ghw.PCIDevice) (string, error)
}

type HardwarePlatform struct{}

func NewHardwarePlatform() *HardwarePlatform {
	return &HardwarePlatform{}
}

func (hp *HardwarePlatform) PciDevices() ([]*ghw.PCIDevice, error) {
	pciInfo, err := ghw.PCI()
	if err != nil {
		return nil, err
	}
	return pciInfo.Devices, nil
}

func (hp *HardwarePlatform) NetDevs() ([]*ghw.NIC, error) {
	netInfo, err := ghw.Network()
	if err != nil {
		return nil, err
	}
	return netInfo.NICs, nil
}

func (hp *HardwarePlatform) Product() (*ghw.ProductInfo, error) {
	return ghw.Product()
}

func (hp *HardwarePlatform) ReadDeviceSerialNumber(pciDevice *ghw.PCIDevice) (string, error) {
	if pciDevice == nil {
		return "", fmt.Errorf("nil PCI device provided")
	}

	devicePath := filepath.Join("/sys/bus/pci/devices", pciDevice.Address, "config")

	file, err := os.Open(devicePath)
	if err != nil {
		return "", fmt.Errorf("failed to open config space: %v", err)
	}
	defer file.Close()

	// Seek to offset 0x150
	// Capabilities: [150] Device Serial Number 88-dc-97-ff-ff-44-24-8b
	const serialOffset = 0x150
	_, err = file.Seek(serialOffset, 0)
	if err != nil {
		return "", fmt.Errorf("seek error to device serial number: %v", err)
	}

	// Read 8 bytes (64-bit serial number) - E.g. 88-dc-97-ff-ff-44-24-8b
	buf := make([]byte, 8)
	_, err = file.Read(buf)
	if err != nil {
		return "", fmt.Errorf("read error from device serial number: : %v", err)
	}

	// Convert raw bytes to hex string
	serialHex := hex.EncodeToString(buf)
	return serialHex, nil
}

type FakePlatform struct {
	platformName string
	devices      []*ghw.PCIDevice
	netdevs      []*ghw.NIC
	mu           sync.Mutex
}

func NewFakePlatform(platformName string) *FakePlatform {
	return &FakePlatform{
		platformName: platformName,
		devices:      make([]*ghw.PCIDevice, 0),
	}
}

func (p *FakePlatform) PciDevices() ([]*ghw.PCIDevice, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.devices, nil
}

func (p *FakePlatform) NetDevs() ([]*ghw.NIC, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.netdevs, nil
}

func (p *FakePlatform) Product() (*ghw.ProductInfo, error) {
	return &ghw.ProductInfo{
		Name: p.platformName,
	}, nil
}

func (p *FakePlatform) ReadDeviceSerialNumber(pciDevice *ghw.PCIDevice) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if pciDevice == nil {
		return "", fmt.Errorf("nil PCI device provided")
	}

	//TODO: Implement a more realistic serial number generation
	return "FAKE-SERIAL-" + pciDevice.Address, nil
}

func (p *FakePlatform) RemoveAllPciDevices() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.devices = make([]*ghw.PCIDevice, 0)
}
