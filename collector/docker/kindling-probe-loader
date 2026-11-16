#!/bin/sh

load_kernel_probe() {
	if ! hash lsmod > /dev/null 2>&1; then
		echo "This program requires lsmod"
		exit 1
	fi

	if ! hash modprobe > /dev/null 2>&1; then
		echo "This program requires modprobe"
		exit 1
	fi

	if ! hash rmmod > /dev/null 2>&1; then
		echo "This program requires rmmod"
		exit 1
	fi

	echo "* Unloading ${PROBE_NAME}, if present"
	rmmod "${PROBE_NAME}" 2>/dev/null
	WAIT_TIME=0
	KMOD_NAME=$(echo "${PROBE_NAME}" | tr "-" "_")
	while lsmod | grep "${KMOD_NAME}" > /dev/null 2>&1 && [ $WAIT_TIME -lt $MAX_RMMOD_WAIT ]; do
		if rmmod "${PROBE_NAME}" 2>/dev/null; then
			echo "* Unloading ${PROBE_NAME} succeeded after ${WAIT_TIME}s"
			break
		fi
		((++WAIT_TIME))
		if (( $WAIT_TIME % 5 == 0 )); then
			echo "* ${PROBE_NAME} still loaded, waited ${WAIT_TIME}s (max wait ${MAX_RMMOD_WAIT}s)"
		fi
		sleep 1
	done

	if lsmod | grep "${KMOD_NAME}" > /dev/null 2>&1; then
		echo "* ${PROBE_NAME} seems to still be loaded, hoping the best"
		exit 0
	fi

	echo "* Trying to find precompiled ${PROBE_NAME} for ${KERNEL_RELEASE}"

	local KINDLING_PROBE_FILENAME="${KERNEL_RELEASE}.ko"
  local APPROXIMATE_PROBE_FILENAME="${APPROXIMATE_KERNEL_VERSION}.ko"
	if [ -f "/opt/.kindling/${KINDLING_PROBE_FILENAME}" ]; then
		echo "* Found precompiled module at /opt/.kindling/${KINDLING_PROBE_FILENAME}, loading module"
		if insmod "/opt/.kindling/${KINDLING_PROBE_FILENAME}" > /dev/null 2>&1; then
		  create_kernel_support_file
			exit $?
		fi
	elif [ -f "/opt/.kindling/${APPROXIMATE_PROBE_FILENAME}" ]; then
		echo "* Found precompiled module with approximate version at /opt/.kindling/${APPROXIMATE_PROBE_FILENAME}, loading module"
		if insmod "/opt/.kindling/${APPROXIMATE_PROBE_FILENAME}" > /dev/null 2>&1; then
		  create_kernel_support_file
		  exit $?
	  fi
	else
	  echo "* Error: Precompiled module at /opt/.kindling/ is not found, and the agent will not work as expected"
	fi
}

load_bpf_probe() {
	echo "* Mounting debugfs"

	if [ ! -d /sys/kernel/debug/tracing ]; then
		mount -t debugfs nodev /sys/kernel/debug
	fi

	local BPF_PROBE_FILENAME="${KERNEL_RELEASE}.o"

	if [ -f "/opt/.kindling/${BPF_PROBE_FILENAME}" ]; then
		if [ ! -f /proc/sys/net/core/bpf_jit_enable ]; then
			echo "**********************************************************"
			echo "** BPF doesn't have JIT enabled, performance might be   **"
			echo "** degraded. Please ensure to run on a kernel with      **"
			echo "** CONFIG_BPF_JIT enabled and/or use --net=host if      **"
			echo "** running inside a container.                          **"
			echo "**********************************************************"
		fi

		echo "* BPF probe located, it's now possible to start kindling"

		ln -sf "/opt/.kindling/${BPF_PROBE_FILENAME}" "/opt/probe.o"
		create_kernel_support_file
		exit $?
	else
		echo "* Failure to find a BPF probe, try to load kernel probe"
		load_kernel_probe
		exit $?
	fi
}

create_kernel_support_file() {
  echo "* Load probe succeeded, and will create /opt/kernel-support for kubernetes"
  touch /opt/kernel-support
}

ARCH=$(uname -m)
KERNEL_RELEASE=$(uname -r)
APPROXIMATE_KERNEL_VERSION=$(echo "$KERNEL_RELEASE" | awk -F. '{print $1"."$2"."$3"."$(NF-1)"."$NF}')
SCRIPT_NAME=$(basename "${0}")
PROBE_NAME="kindling-falcolib-probe"
MAX_RMMOD_WAIT=60

tar -zxvf /app/kindling-falcolib-probe.tar.gz -C /opt
if [ -d /opt/.kindling ]; then
  mv -f /opt/kindling-falcolib-probe/* /opt/.kindling/
else
  mv /opt/kindling-falcolib-probe/ /opt/.kindling/
fi

if [ -d /opt/kindling-extra-probe ]; then
	/bin/cp -rf /opt/kindling-extra-probe/* /opt/.kindling/
fi

if [ "$(id -u)" != 0 ]; then
	echo "Installer must be run as root (or with sudo)."
	exit 1
fi
# PREFER_KERNEL_PROBE不为0，加载kernel_probe
if [ ${PREFER_KERNEL_PROBE} ] && [ "${PREFER_KERNEL_PROBE}" != 0 ] ; then
	load_kernel_probe
else
	load_bpf_probe
fi
