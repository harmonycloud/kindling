// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// Copied from github.com/DataDog/datadog-agent/pkg/util/kernel and pkg/process/util

package internal

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime"
	"strconv"

	"github.com/DataDog/ebpf"
	"github.com/vishvananda/netns"
	"golang.org/x/sys/unix"
)

// GetNetNamespaces returns a list of network namespaces on the machine. The caller
// is responsible for calling Close() on each of the returned NsHandle's.
func GetNetNamespaces(procRoot string) ([]netns.NsHandle, error) {
	var nss []netns.NsHandle
	seen := make(map[string]interface{})
	err := WithAllProcs(procRoot, func(pid int) error {
		ns, err := netns.GetFromPath(path.Join(procRoot, fmt.Sprintf("%d/ns/net", pid)))
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) && !errors.Is(err, unix.ENOENT) {
				log.Printf("error while reading %s: %s", path.Join(procRoot, fmt.Sprintf("%d/ns/net", pid)), err)
			}
			return nil
		}

		uid := ns.UniqueId()
		if _, ok := seen[uid]; ok {
			ns.Close()
			return nil
		}

		seen[uid] = struct{}{}
		nss = append(nss, ns)
		return nil
	})

	if err != nil {
		// close all the accumulated ns handles
		for _, ns := range nss {
			ns.Close()
		}

		return nil, err
	}

	return nss, nil
}

// WithAllProcs will execute `fn` for every pid under procRoot. `fn` is
// passed the `pid`. If `fn` returns an error the iteration aborts,
// returning the last error returned from `fn`.
func WithAllProcs(procRoot string, fn func(int) error) error {
	files, err := ioutil.ReadDir(procRoot)
	if err != nil {
		return err
	}

	for _, f := range files {
		if !f.IsDir() || f.Name() == "." || f.Name() == ".." {
			continue
		}

		var pid int
		if pid, err = strconv.Atoi(f.Name()); err != nil {
			continue
		}

		if err = fn(pid); err != nil {
			return err
		}
	}

	return nil
}

// GetRootNetNamespace gets the root network namespace
func GetRootNetNamespace(procRoot string) (netns.NsHandle, error) {
	return GetNetNamespaceFromPid(procRoot, 1)
}

// GetNetNamespaceFromPid gets the network namespace for a given `pid`
func GetNetNamespaceFromPid(procRoot string, pid int) (netns.NsHandle, error) {
	return netns.GetFromPath(path.Join(procRoot, fmt.Sprintf("%d/ns/net", pid)))
}

// WithNS executes the given function in the given network namespace, and then
// switches back to the previous namespace.
func WithNS(procRoot string, ns netns.NsHandle, fn func() error) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	prevNS, err := netns.Get()
	if err != nil {
		return err
	}

	if ns.Equal(prevNS) {
		return fn()
	}

	if err := netns.Set(ns); err != nil {
		return err
	}

	fnErr := fn()
	nsErr := netns.Set(prevNS)
	if fnErr != nil {
		return fnErr
	}
	return nsErr
}

// WithRootNS executes a function within root network namespace and then switch back
// to the previous namespace. If the thread is already in the root network namespace,
// the function is executed without calling SYS_SETNS.
func WithRootNS(procRoot string, fn func() error) error {
	rootNS, err := GetRootNetNamespace(procRoot)
	if err != nil {
		return err
	}

	return WithNS(procRoot, rootNS, fn)
}

// Version is a numerical representation of a kernel version
type Version uint32

var hostVersion Version

// String returns a string representing the version in x.x.x format
func (v Version) String() string {
	a, b, c := v>>16, v>>8&0xff, v&0xff
	return fmt.Sprintf("%d.%d.%d", a, b, c)
}

// HostVersion returns the running kernel version of the host
func HostVersion() (Version, error) {
	if hostVersion != 0 {
		return hostVersion, nil
	}
	kv, err := ebpf.CurrentKernelVersion()
	if err != nil {
		return 0, err
	}
	hostVersion = Version(kv)
	return hostVersion, nil
}

// VersionCode returns a Version computed from the individual parts of a x.x.x version
func VersionCode(major, minor, patch byte) Version {
	// KERNEL_VERSION(a,b,c) = (a << 16) + (b << 8) + (c)
	// Per https://github.com/torvalds/linux/blob/db7c953555388571a96ed8783ff6c5745ba18ab9/Makefile#L1250
	return Version((uint32(major) << 16) + (uint32(minor) << 8) + uint32(patch))
}
