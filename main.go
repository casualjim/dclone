// Copyright © 2016 Ivan Porto Carrero
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strconv"
	"strings"

	units "github.com/docker/go-units"
	docker "github.com/fsouza/go-dockerclient"
)

type writer interface {
	WriteTo(io.Writer) error
}

/*
Not supported at this stage

      --cpu-percent int             CPU percent (Windows only)
      --detach-keys string          Override the key sequence for detaching a container
      --expose value                Expose a port or a range of ports (default [])
      --health-cmd string           Command to run to check health
      --health-interval duration    Time between running the check
      --health-retries int          Consecutive failures needed to report unhealthy
      --health-timeout duration     Maximum time to allow one check to run
      --io-maxbandwidth string      Maximum IO bandwidth limit for the system drive (Windows only)
      --io-maxiops uint             Maximum IOps limit for the system drive (Windows only)
      --ip string                   Container IPv4 address (e.g. 172.30.100.104)
      --ip6 string                  Container IPv6 address (e.g. 2001:db8::33)
      --isolation string            Container isolation technology
      --link-local-ip value         Container IPv4/IPv6 link-local addresses (default [])
      --network-alias value         Add network-scoped alias for the container (default [])
      --no-healthcheck              Disable any container-specified HEALTHCHECK
      --runtime string              Runtime to use for this container
      --sig-proxy                   Proxy received signals to the process (default true)
*/

func main() {
	log.SetFlags(0)

	dc, err := docker.NewClientFromEnv()
	if err != nil {
		log.Fatalln(err)
	}

	if len(os.Args) < 2 {
		log.Fatalln("You need to specify a container name or id")
	}

	ctr, err := dc.InspectContainer(os.Args[1])
	if err != nil {
		log.Fatalln(err)
	}
	img, err := dc.InspectImage(ctr.Image)
	if err != nil {
		log.Fatalln(err)
	}
	// pretty.Println(ctr)
	// pretty.Println(img)

	parts := []string{"docker", "run"}
	if ctr.HostConfig.AutoRemove {
		parts = append(parts, "--rm")
	}
	if ctr.Config.OpenStdin {
		parts = append(parts, "--interactive")
	}
	if ctr.Config.Tty {
		parts = append(parts, "--tty")
	}
	if !ctr.Config.AttachStdin || !ctr.Config.AttachStdout || !ctr.Config.AttachStderr {
		if !ctr.Config.AttachStdin && !ctr.Config.AttachStdout && !ctr.Config.AttachStderr {
			parts = append(parts, "--detach")
		} else {
			if ctr.Config.AttachStdin {
				parts = append(parts, "--attach", "stdin")
			}
			if ctr.Config.AttachStdout {
				parts = append(parts, "--attach", "stdout")
			}
			if ctr.Config.AttachStderr {
				parts = append(parts, "--attach", "stderr")
			}
		}
	}

	if ctr.Name != "" {
		parts = append(parts, "--name", path.Base(ctr.Name))
	}
	if ctr.Config.Hostname != "" && !strings.HasPrefix(ctr.ID, ctr.Config.Hostname) {
		parts = append(parts, "--hostname", ctr.Config.Hostname)
	}
	if ctr.HostConfig.NetworkMode != "default" {
		parts = append(parts, "--network", ctr.HostConfig.NetworkMode)
	}
	if ctr.HostConfig.Memory != 0 {
		parts = append(parts, "--memory", units.HumanSize(float64(ctr.HostConfig.Memory)))
	}
	if ctr.HostConfig.MemoryReservation != 0 {
		parts = append(parts, "--memory-reservation", units.HumanSize(float64(ctr.HostConfig.MemoryReservation)))
	}
	if ctr.HostConfig.MemorySwap != 0 {
		parts = append(parts, "--memory-swap", units.HumanSize(float64(ctr.HostConfig.MemorySwap)))
	}
	if ctr.HostConfig.MemorySwappiness != -1 {
		parts = append(parts, "--memory-swappiness", strconv.FormatInt(ctr.HostConfig.MemorySwappiness, 10))
	}
	if ctr.HostConfig.KernelMemory != 0 {
		parts = append(parts, "--kernel-memory", units.HumanSize(float64(ctr.HostConfig.KernelMemory)))
	}
	if ctr.HostConfig.CPUPeriod != 0 {
		parts = append(parts, "--cpu-period", strconv.FormatInt(ctr.HostConfig.CPUPeriod, 10))
	}
	if ctr.HostConfig.CPUQuota != 0 {
		parts = append(parts, "--cpu-quota", strconv.FormatInt(ctr.HostConfig.CPUQuota, 10))
	}
	if ctr.HostConfig.CPUShares != 0 {
		parts = append(parts, "--cpu-shares", strconv.FormatInt(ctr.HostConfig.CPUShares, 10))
	}
	if ctr.HostConfig.CPUSetCPUs != "" {
		parts = append(parts, "--cpuset-cpus", ctr.HostConfig.CPUSetCPUs)
	}
	if ctr.HostConfig.CPUSetMEMs != "" {
		parts = append(parts, "--cpuset-mems", ctr.HostConfig.CPUSetMEMs)
	}
	if ctr.HostConfig.Privileged {
		parts = append(parts, "--privileged")
	}
	if ctr.HostConfig.PublishAllPorts {
		parts = append(parts, "--publish-all")
	}
	if ctr.HostConfig.ReadonlyRootfs {
		parts = append(parts, "--read-only")
	}
	if ctr.Config.MacAddress != "" {
		parts = append(parts, "--mac-address", ctr.Config.MacAddress)
	}
	if len(ctr.Config.Entrypoint) > 0 {
		parts = append(parts, "--entrypoint", fmt.Sprintf("'%s'", strings.Join(ctr.Config.Entrypoint, " ")))
	}

	if ctr.Config.WorkingDir != "" {
		parts = append(parts, "--workdir", ctr.Config.WorkingDir)
	}
	if len(ctr.HostConfig.Binds) > 0 {
		for _, b := range ctr.HostConfig.Binds {
			parts = append(parts, "--volume", b)
		}
	}
	if ctr.HostConfig.VolumeDriver != "" {
		parts = append(parts, "--volume-driver", ctr.HostConfig.VolumeDriver)
	}
	for _, vf := range ctr.HostConfig.VolumesFrom {
		parts = append(parts, "--volumes-from", vf)
	}

	for _, ca := range ctr.HostConfig.CapAdd {
		parts = append(parts, "--cap-add", ca)
	}

	for _, cd := range ctr.HostConfig.CapDrop {
		parts = append(parts, "--cap-drop", cd)
	}

	for _, ga := range ctr.HostConfig.GroupAdd {
		parts = append(parts, "--group-add", ga)
	}

	for _, l := range ctr.HostConfig.Links {
		parts = append(parts, "--link", l)
	}

	for _, d := range ctr.HostConfig.DNS {
		parts = append(parts, "--dns", d)
	}

	for _, do := range ctr.HostConfig.DNSOptions {
		parts = append(parts, "--dns-opt", do)
	}

	for _, ds := range ctr.HostConfig.DNSSearch {
		parts = append(parts, "--dns-search", ds)
	}

	if ctr.HostConfig.ContainerIDFile != "" {
		parts = append(parts, "--cidfile", ctr.HostConfig.ContainerIDFile)
	}

	if ctr.HostConfig.BlkioWeight > 0 {
		parts = append(parts, "--blkio-weight", strconv.FormatInt(ctr.HostConfig.BlkioWeight, 10))
	}
	for k, v := range ctr.Config.Labels {
		parts = append(parts, "--label", fmt.Sprintf("%s=%s", k, v))
	}
	for _, bwd := range ctr.HostConfig.BlkioWeightDevice {
		var res []string
		if bwd.Path != "" {
			res = append(res, bwd.Path)
			if bwd.Weight != "" {
				res = append(res, bwd.Weight)
			}
		}
		if len(res) > 0 {
			parts = append(parts, "--blkio-weight-device", strings.Join(res, ":"))
		}
	}

	for _, h := range ctr.HostConfig.ExtraHosts {
		parts = append(parts, "--add-host", h)
	}

	if ctr.HostConfig.CgroupParent != "" {
		parts = append(parts, "--cgroup-parent", ctr.HostConfig.CgroupParent)
	}

	if ctr.HostConfig.RestartPolicy.Name != "no" {
		pol := ctr.HostConfig.RestartPolicy.Name
		if pol == "on-failure" {
			pol += ":" + strconv.Itoa(ctr.HostConfig.RestartPolicy.MaximumRetryCount)
		}
		parts = append(parts, "--restart", pol)
	}

	if ctr.HostConfig.OOMKillDisable {
		parts = append(parts, "--oom-kill-disable")
	}
	if ctr.HostConfig.OomScoreAdj != 0 {
		parts = append(parts, "--oom-score-adj", strconv.Itoa(ctr.HostConfig.OomScoreAdj))
	}

	for _, ul := range ctr.HostConfig.Ulimits {
		uu := &units.Ulimit{
			Hard: ul.Hard,
			Name: ul.Name,
			Soft: ul.Soft,
		}
		parts = append(parts, "--ulimit", uu.String())
	}

	for k, tm := range ctr.HostConfig.Tmpfs {
		r := k
		if tm != "" {
			r += ":" + tm
		}
		parts = append(parts, "--tmpfs", r)
	}

	if ctr.HostConfig.UsernsMode != "" {
		parts = append(parts, "--userns", ctr.HostConfig.UsernsMode)
	}
	if ctr.HostConfig.UTSMode != "" {
		parts = append(parts, "--uts", ctr.HostConfig.UTSMode)
	}
	if ctr.Config.User != "" {
		parts = append(parts, "--user", ctr.Config.User)
	}
	if ctr.HostConfig.PidMode != "" {
		parts = append(parts, "--pid", ctr.HostConfig.PidMode)
	}
	if ctr.HostConfig.PidsLimit > 0 {
		parts = append(parts, "--pids-limit", strconv.FormatInt(ctr.HostConfig.PidsLimit, 10))
	}
	if ctr.HostConfig.IpcMode != "" {
		parts = append(parts, "--ipc", ctr.HostConfig.IpcMode)
	}
	if ctr.Config.StopSignal != "" {
		parts = append(parts, "--stop-signal", ctr.Config.StopSignal)
	}

	for nm, so := range ctr.HostConfig.StorageOpt {
		parts = append(parts, "--storage-opt", fmt.Sprintf("%s=%s", nm, so))
	}

	if ctr.HostConfig.ShmSize != 67108864 {
		parts = append(parts, "--shm-size", units.HumanSize(float64(ctr.HostConfig.ShmSize)))
	}
	for k, v := range ctr.HostConfig.Sysctls {
		parts = append(parts, "--sysctl", fmt.Sprintf("%s=%s", k, v))
	}
	if ctr.HostConfig.LogConfig.Type != "json-file" || len(ctr.HostConfig.LogConfig.Config) > 0 {
		parts = append(parts, "--log-driver", ctr.HostConfig.LogConfig.Type)
		for k, v := range ctr.HostConfig.LogConfig.Config {
			parts = append(parts, "--log-opt", fmt.Sprintf("%s=%s", k, v))
		}
	}
	for _, so := range ctr.HostConfig.SecurityOpt {
		parts = append(parts, "--security-opt", so)
	}
	for _, dev := range ctr.HostConfig.Devices {
		r := dev.PathOnHost
		if dev.PathInContainer != "" {
			r += ":" + dev.PathInContainer
			if dev.CgroupPermissions != "" {
				r += ":" + dev.CgroupPermissions
			}
		}
		parts = append(parts, "--device", r)
	}
	for _, bl := range ctr.HostConfig.BlkioDeviceReadBps {
		r := bl.Path
		if bl.Rate != "" {
			r += ":" + bl.Rate
		}
		parts = append(parts, "--device-read-bps", r)
	}
	for _, bl := range ctr.HostConfig.BlkioDeviceWriteBps {
		r := bl.Path
		if bl.Rate != "" {
			r += ":" + bl.Rate
		}
		parts = append(parts, "--device-write-bps", r)
	}
	for _, bl := range ctr.HostConfig.BlkioDeviceReadIOps {
		r := bl.Path
		if bl.Rate != "" {
			r += ":" + bl.Rate
		}
		parts = append(parts, "--device-read-iops", r)
	}
	for _, bl := range ctr.HostConfig.BlkioDeviceWriteIOps {
		r := bl.Path
		if bl.Rate != "" {
			r += ":" + bl.Rate
		}
		parts = append(parts, "--device-write-iops", r)
	}

	if !ctr.HostConfig.PublishAllPorts && len(ctr.HostConfig.PortBindings) > 0 {
		for dp, pbs := range ctr.HostConfig.PortBindings {
			for _, pb := range pbs {
				var rpb []string
				if pb.HostIP != "" {
					rpb = append(rpb, pb.HostIP)
				}
				if pb.HostPort != "" {
					rpb = append(rpb, pb.HostPort)
				}
				rpbs := strings.Join(rpb, ":")
				var dpb []string
				if rpbs != "" {
					dpb = append(dpb, rpbs)
				}
				dpbs := dp.Port()
				if dp.Proto() != "tcp" {
					dpbs += "/" + dp.Proto()
				}
				dpb = append(dpb, dpbs)
				parts = append(parts, "--publish", strings.Join(dpb, ":"))
			}
		}
	}

	if ctr.Config.Image != "" {
		parts = append(parts, ctr.Config.Image)
	}
	var hasDiffArg bool
	for _, p := range ctr.Config.Cmd {
		var hasArg bool
		for _, c := range img.Config.Cmd {
			if p == c {
				hasArg = true
				break
			}
		}
		if !hasArg {
			hasDiffArg = true
			break
		}
	}
	if hasDiffArg {
		parts = append(parts, ctr.Config.Cmd...)
	}
	fmt.Println(strings.Join(parts, " "))
}
