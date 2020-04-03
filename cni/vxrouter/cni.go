package main

import (
	"bufio"
	"encoding/json"
	"net"
	"os"

	"github.com/TrilliumIT/vxrouter"
	"github.com/TrilliumIT/vxrouter/cni"
	"github.com/TrilliumIT/vxrouter/host"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

func main() {
	//populate cni config from standard input
	scanner := bufio.NewScanner(os.Stdin)
	var confBytes []byte
	for scanner.Scan() {
		confBytes = append(confBytes, scanner.Bytes()...)
	}
	if len(confBytes) == 0 {
		log.Fatal("No bytes sent on stdin. This program must be envoked by a CNI compatible container runtime.")
	}

	var conf *cni.Config
	err := json.Unmarshal(confBytes, conf)
	if err != nil {
		log.Fatal("failed to parse json config on stdin")
	}

	//read environment for required info
	command := os.Getenv("CNI_COMMAND")
	//cid := os.Getenv("CNI_CONTAINERID")
	netns := os.Getenv("CNI_NETNS")
	ifname := os.Getenv("CNI_IFNAME")
	//args := os.Getenv("CNI_ARGS")
	//path := os.Getenv("CNI_PATH")

	//container specific variables
	//name of the vxlan for this container
	vxlan, ok := conf.Args.Attributes["vxrouter/network"].(string)
	if !ok {
		log.Fatal("vxlan name to connect to was not specified in 'args.attributes.vxrouter/network'")
	}
	//find config for select vxlan
	var vxlConf *cni.VxlanConfig
	for _, v := range conf.Vxlans {
		if vxlan == v.Name {
			vxlConf = v
		}
	}
	if vxlConf == nil {
		log.Fatalf("vxlan name %v has no corresponding configuration")
	}
	//requested address for this container
	cip, ok := conf.Args.Attributes["vxrouter/address"].(string)
	if !ok {
		log.Infof("no specific address requested, will use a dynamically chosen address")
		cip = ""
	}
	var ip net.IP
	if cip != "" {
		ip = net.ParseIP(cip)
		if ip == nil {
			log.Warningf("requested address %v could not be parsed, using dynamically chosen address")
		}
	}

	//host specific variables
	//get the hosts gateway address and subnet for this vxlan
	gw, net, err := net.ParseCIDR(vxlConf.Cidr)
	if err != nil {
		log.Fatalf("could not parse cidr %v for vxlan %v", vxlConf.Cidr, vxlan)
	}
	net.IP = gw

	//open the namespace file for this containers netns
	file, err := os.Open(netns)
	if err != nil {
		log.Fatalf("failed to open netns path %v", netns)
	}

	switch command {
	case "ADD":
		//get host interface
		hi, err := host.GetOrCreateInterface(vxlan, net, vxlConf.Options)
		if err != nil {
			log.Fatalf("failed to get or create vxlan interface %v", vxlan)
		}
		//create new container interface
		cmvl, err := hi.CreateMacvlan(ifname)
		if err != nil {
			log.Fatalf("failed to create container macvlan %v", ifname)
		}
		//select an IP address (or verify requested address is available)
		addr, err := hi.SelectAddress(ip, vxrouter.DefaultReqAddrSleepTime, vxrouter.DefaultReqAddrTimeout, vxlConf.ExcludeFirst, vxlConf.ExcludeLast)
		if err != nil {
			log.Fatalf("failed to select an address for this container interface")
		}
		//add selected address to our container interface
		err = cmvl.AddAddress(addr)
		if err != nil {
			log.Fatalf("failed to add selected address to container interface")
		}

		//get netlink link object for our new cmvl
		link, err := netlink.LinkByName(ifname)
		if err != nil {
			log.Fatalf("failed to get link for %v", ifname)
		}
		//put our cmvl into the provided netns
		err = netlink.LinkSetNsFd(link, int(file.Fd()))
		if err != nil {
			log.Fatalf("failed to insert interface %v into namespace %v", ifname, netns)
		}
	case "DEL":
	case "CHECK":
	case "VERSION":
	default:
		log.Fatal("This program must be envoked by a CNI compatible container runtime.")
	}

}
