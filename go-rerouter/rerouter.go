package main

//go:generate sh -c "rm -f *_bpfel.go *_bpfeb.go"
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -target amd64 -type Config rerouter rerouter.ebpf.c -- -g -I/usr/include/i386-linux-gnu -I/usr/i686-linux-gnu/include

import (
	"fmt"
	"log"
	"os"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/rlimit"
)

const (
	CGROUP_PATH     = "/sys/fs/cgroup" // Root cgroup path
	TUNNEL_PORT     = 18000            // Port where the tunnel (in go or c) listens
	REAL_PROXY_PORT = 8080             // Port where the real proxy (cntlm) listens
)

func addBoundPidForPort(port uint16, boundPidsMap *ebpf.Map) error {
	pid, err := getPidFromPort(port)
	if err != nil {
		return err
	}

	// if we found the PID listening on this port, update the the `bound_pid` map,
	// since that means the proxy already called bind() so it won't be caught by our eBPF
	log.Printf("Process for port %d has PID %d, will ignore any connection from it", port, pid)

	err = boundPidsMap.Update(&pid, &port, ebpf.UpdateAny)
	if err != nil {
		return fmt.Errorf("failed to update boundPids map: %v", err)
	}

	return nil
}

func addPortToWhitelist(port uint16, configMap *ebpf.Map, whitelist *ebpf.Map) error {
	var key uint32 = 0
	var config rerouterConfig

	err := configMap.Lookup(&key, &config)
	if err != nil {
		return err
	}

	if config.WhitelistCount >= uint8(whitelist.MaxEntries()) {
		return fmt.Errorf("reached maximum whitelist capacity (%d), can't add any more", uint8(whitelist.MaxEntries()))
	}

	nextIdx := uint32(config.WhitelistCount)
	err = whitelist.Update(&nextIdx, &port, ebpf.UpdateAny)
	if err != nil {
		return err
	}

	config.WhitelistCount++

	err = configMap.Update(&key, &config, ebpf.UpdateExist)
	if err != nil {
		return err
	}

	return nil
}

func whitelistPort(port uint16, maps rerouterMaps) error {
	err := addPortToWhitelist(port, maps.MapConfig, maps.MapWhitelistPorts)
	if err != nil {
		return err
	}

	err = addBoundPidForPort(port, maps.MapBoundPids)
	if err != nil {
		return fmt.Errorf("tried to whitelist port '%d' but: %v. We'll catch the PID with bind() hook", port, err)
	}

	return nil
}

func main() {
	// Remove resource limits for kernels <5.11.
	if err := rlimit.RemoveMemlock(); err != nil {
		log.Print("Removing memlock:", err)
	}

	// Load the compiled eBPF ELF and load it into the kernel
	// NOTE: we could also pin the eBPF program
	var objs rerouterObjects
	if err := loadRerouterObjects(&objs, nil); err != nil {
		log.Fatalf("Error loading eBPF program: %v", err)
	}
	defer objs.Close()

	// Attach eBPF programs to the root cgroup
	connect4Link, err := link.AttachCgroup(link.CgroupOptions{
		Path:    CGROUP_PATH,
		Attach:  ebpf.AttachCGroupInet4Connect,
		Program: objs.CgConnect4,
	})
	if err != nil {
		log.Fatalf("Attaching CgConnect4 program to Cgroup: %v", err)
	}
	defer connect4Link.Close()

	sockopsLink, err := link.AttachCgroup(link.CgroupOptions{
		Path:    CGROUP_PATH,
		Attach:  ebpf.AttachCGroupSockOps,
		Program: objs.CgSockOps,
	})
	if err != nil {
		log.Fatalf("Attaching CgSockOps program to Cgroup: %v", err)
	}
	defer sockopsLink.Close()

	sockoptLink, err := link.AttachCgroup(link.CgroupOptions{
		Path:    CGROUP_PATH,
		Attach:  ebpf.AttachCGroupGetsockopt,
		Program: objs.CgSockOpt,
	})
	if err != nil {
		log.Fatalf("Attaching CgSockOpt program to Cgroup: %v", err)
	}
	defer sockoptLink.Close()

	boundpidsLink, err := link.AttachCgroup(link.CgroupOptions{
		Path:    CGROUP_PATH,
		Attach:  ebpf.AttachCGroupInet4PostBind,
		Program: objs.CgPostBind4,
	})
	if err != nil {
		log.Fatalf("Attaching CgPostBind4 program to Cgroup: %v", err)
	}
	defer boundpidsLink.Close()

	cloneProbeLink, err := link.Kretprobe("sys_clone", objs.ProbeClone, nil)
	if err != nil {
		log.Fatalf("Attaching clone probe: %v", err)
	}
	defer cloneProbeLink.Close()

	clone3ProbeLink, err := link.Kretprobe("sys_clone3", objs.ProbeClone3, nil)
	if err != nil {
		log.Fatalf("Attaching clone3 probe: %v", err)
	}
	defer clone3ProbeLink.Close()

	// Update the rerouter map with the proxy server configuration, because we need to know the proxy server PID in order
	// to filter out eBPF events generated by the proxy server itself so it would not proxy its own packets in a loop.
	var key uint32 = 0
	config := rerouterConfig{
		ProxyPort:     TUNNEL_PORT,
		ProxyPid:      uint64(os.Getpid()),
		RealProxyPort: REAL_PROXY_PORT,
		// RealProxyPid:  realProxyPid,
		WhitelistCount: 0,
	}
	err = objs.rerouterMaps.MapConfig.Update(&key, &config, ebpf.UpdateAny)
	if err != nil {
		log.Fatalf("Failed to update proxyMaps map: %v", err)
	}

	err = whitelistPort(REAL_PROXY_PORT, objs.rerouterMaps)
	if err == nil {
		log.Print("Added proxy to rerouter whitelist successfully")
	} else {
		log.Print(err)
	}

	whitelistPort(TUNNEL_PORT, objs.rerouterMaps)
	if err == nil {
		log.Print("Added tunnel to rerouter whitelist successfully")
	} else {
		log.Print(err)
	}

	log.Printf("eBPF rerouter setup finished, redirecting all requests to %d", TUNNEL_PORT)
	fmt.Print("Press Enter to exit")
	fmt.Scanln()
}
