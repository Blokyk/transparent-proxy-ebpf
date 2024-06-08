// Code generated by bpf2go; DO NOT EDIT.
//go:build 386 || amd64 || arm || arm64 || loong64 || mips64le || mipsle || ppc64le || riscv64

package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"

	"github.com/cilium/ebpf"
)

type proxyConfig struct {
	ProxyPort uint16
	_         [6]byte
	ProxyPid  uint64
}

type proxySocket struct {
	SrcAddr uint32
	SrcPort uint16
	_       [2]byte
	DstAddr uint32
	DstPort uint16
	_       [2]byte
}

// loadProxy returns the embedded CollectionSpec for proxy.
func loadProxy() (*ebpf.CollectionSpec, error) {
	reader := bytes.NewReader(_ProxyBytes)
	spec, err := ebpf.LoadCollectionSpecFromReader(reader)
	if err != nil {
		return nil, fmt.Errorf("can't load proxy: %w", err)
	}

	return spec, err
}

// loadProxyObjects loads proxy and converts it into a struct.
//
// The following types are suitable as obj argument:
//
//	*proxyObjects
//	*proxyPrograms
//	*proxyMaps
//
// See ebpf.CollectionSpec.LoadAndAssign documentation for details.
func loadProxyObjects(obj interface{}, opts *ebpf.CollectionOptions) error {
	spec, err := loadProxy()
	if err != nil {
		return err
	}

	return spec.LoadAndAssign(obj, opts)
}

// proxySpecs contains maps and programs before they are loaded into the kernel.
//
// It can be passed ebpf.CollectionSpec.Assign.
type proxySpecs struct {
	proxyProgramSpecs
	proxyMapSpecs
}

// proxySpecs contains programs before they are loaded into the kernel.
//
// It can be passed ebpf.CollectionSpec.Assign.
type proxyProgramSpecs struct {
	CgConnect4 *ebpf.ProgramSpec `ebpf:"cg_connect4"`
	CgSockOps  *ebpf.ProgramSpec `ebpf:"cg_sock_ops"`
	CgSockOpt  *ebpf.ProgramSpec `ebpf:"cg_sock_opt"`
}

// proxyMapSpecs contains maps before they are loaded into the kernel.
//
// It can be passed ebpf.CollectionSpec.Assign.
type proxyMapSpecs struct {
	MapConfig *ebpf.MapSpec `ebpf:"map_config"`
	MapPorts  *ebpf.MapSpec `ebpf:"map_ports"`
	MapSocks  *ebpf.MapSpec `ebpf:"map_socks"`
}

// proxyObjects contains all objects after they have been loaded into the kernel.
//
// It can be passed to loadProxyObjects or ebpf.CollectionSpec.LoadAndAssign.
type proxyObjects struct {
	proxyPrograms
	proxyMaps
}

func (o *proxyObjects) Close() error {
	return _ProxyClose(
		&o.proxyPrograms,
		&o.proxyMaps,
	)
}

// proxyMaps contains all maps after they have been loaded into the kernel.
//
// It can be passed to loadProxyObjects or ebpf.CollectionSpec.LoadAndAssign.
type proxyMaps struct {
	MapConfig *ebpf.Map `ebpf:"map_config"`
	MapPorts  *ebpf.Map `ebpf:"map_ports"`
	MapSocks  *ebpf.Map `ebpf:"map_socks"`
}

func (m *proxyMaps) Close() error {
	return _ProxyClose(
		m.MapConfig,
		m.MapPorts,
		m.MapSocks,
	)
}

// proxyPrograms contains all programs after they have been loaded into the kernel.
//
// It can be passed to loadProxyObjects or ebpf.CollectionSpec.LoadAndAssign.
type proxyPrograms struct {
	CgConnect4 *ebpf.Program `ebpf:"cg_connect4"`
	CgSockOps  *ebpf.Program `ebpf:"cg_sock_ops"`
	CgSockOpt  *ebpf.Program `ebpf:"cg_sock_opt"`
}

func (p *proxyPrograms) Close() error {
	return _ProxyClose(
		p.CgConnect4,
		p.CgSockOps,
		p.CgSockOpt,
	)
}

func _ProxyClose(closers ...io.Closer) error {
	for _, closer := range closers {
		if err := closer.Close(); err != nil {
			return err
		}
	}
	return nil
}

// Do not access this directly.
//
//go:embed proxy_bpfel.o
var _ProxyBytes []byte
