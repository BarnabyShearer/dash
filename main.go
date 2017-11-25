// Mainly copied from https://github.com/google/gopacket/examples/arpscan/arpscan.go
// Copyright 2012 Google, Inc. All rights reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file in the root of the source
// tree.
//
// Needs:
//   sudo setcap cap_net_raw,cap_net_admin=eip dash

package main

import (
	"errors"
	"flag"
	"log"
	"net"
	"os"
	"os/exec"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/mattn/go-shellwords"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return "my string representation"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var macs arrayFlags
var cmds arrayFlags
var last map[string]time.Time

func main() {
	last = make(map[string]time.Time)

	ifaceName := flag.String("iface", "enp3s0", "interface to listen on")
	flag.Var(&macs, "MAC", "MAC address of Dash button to look for (repeat for multiple buttons)")
	flag.Var(&cmds, "cmd", "cmd to run when Dash button pressed (in same order as matching MAC)")

	flag.Parse()

	ifaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}

	for _, iface := range ifaces {
		if *ifaceName == iface.Name {
			if err := scan(&iface); err != nil {
				log.Printf("interface %v: %v", iface.Name, err)
			}
		}
	}

}

func scan(iface *net.Interface) error {
	// We just look for IPv4 addresses, so try to find if the interface has one.
	var addr *net.IPNet
	addrs, err := iface.Addrs()

	if err != nil {
		return err
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok {
			if ip4 := ipnet.IP.To4(); ip4 != nil {
				addr = &net.IPNet{
					IP:   ip4,
					Mask: ipnet.Mask[len(ipnet.Mask)-4:],
				}
				break
			}
		}
	}

	// Sanity-check that the interface has a good address.
	if addr == nil {
		return errors.New("no good IP network found")
	} else if addr.IP[0] == 127 {
		return errors.New("skipping localhost")
	} else if addr.Mask[0] != 0xff || addr.Mask[1] != 0xff {
		return errors.New("mask means network is too large")
	}
	log.Printf("Using network range %v for interface %v", addr, iface.Name)

	// Open up a pcap handle for packet reads/writes.
	handle, err := pcap.OpenLive(iface.Name, 65536, true, pcap.BlockForever)
	if err != nil {
		return err
	}
	defer handle.Close()

	src := gopacket.NewPacketSource(handle, layers.LayerTypeEthernet)
	in := src.Packets()
	for {
		var packet gopacket.Packet
		select {
		case packet = <-in:
			arpLayer := packet.Layer(layers.LayerTypeARP)
			if arpLayer == nil {
				continue
			}
			arp := arpLayer.(*layers.ARP)
			go handleArp(arp)
		}
	}
}

func handleArp(arp *layers.ARP) {
	var cmd []string

	for i, v := range macs {
		if v == net.HardwareAddr(arp.SourceHwAddress).String() {
			log.Printf("IP %v is at %v", net.IP(arp.SourceProtAddress), net.HardwareAddr(arp.SourceHwAddress))
			if last[v].After(time.Now()) {
				log.Printf(" Too fast")
			} else {
				last[v] = time.Now().Local().Add(time.Duration(2) * time.Second)
				cmd, _ = shellwords.Parse(cmds[i])
				break
			}
		}
	}
	if len(cmd) > 0 {
		log.Printf(" running %v", cmd)
		cmd := exec.Command(cmd[0], cmd[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			log.Printf("  Error: %v", err)
		}
	}
}
