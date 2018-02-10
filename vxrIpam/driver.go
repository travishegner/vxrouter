package vxrIpam

import (
	"fmt"
	"net"
	"time"

	"github.com/TrilliumIT/docker-vxrouter/vxrNet"
	"github.com/TrilliumIT/iputil"
	"github.com/docker/go-plugins-helpers/ipam"
	"github.com/vishvananda/netlink"

	log "github.com/Sirupsen/logrus"
)

type Driver struct {
	vxrNet       *vxrNet.Driver
	propTime     time.Duration
	respTime     time.Duration
	excludeFirst int
	excludeLast  int
	log          *log.Entry
}

func NewDriver(vxrNet *vxrNet.Driver, propTime, respTime time.Duration, excludeFirst, excludeLast int) (*Driver, error) {
	d := &Driver{
		vxrNet,
		propTime,
		respTime,
		excludeFirst,
		excludeLast,
		log.WithField("driver", "vxrIpam"),
	}
	return d, nil
}

func (d *Driver) GetCapabilities() (*ipam.CapabilitiesResponse, error) {
	d.log.Debug("GetCapabilites()")
	return &ipam.CapabilitiesResponse{}, nil
}

func (d *Driver) GetDefaultAddressSpaces() (*ipam.AddressSpacesResponse, error) {
	d.log.Debug("GetDefaultAddressSpaces()")
	return &ipam.AddressSpacesResponse{}, nil
}

func (d *Driver) RequestPool(r *ipam.RequestPoolRequest) (*ipam.RequestPoolResponse, error) {
	d.log.WithField("r", r).Debug("RequestPool()")
	return &ipam.RequestPoolResponse{
		PoolID: r.Pool,
		Pool:   r.Pool,
	}, nil
}

func (d *Driver) ReleasePool(r *ipam.ReleasePoolRequest) error {
	d.log.WithField("r", r).Debug("ReleasePoolRequest()")
	return nil
}

func getAddresses(address, subnet string) (*net.IPNet, *net.IPNet, *net.IPNet, error) {
	_, sn, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, nil, nil, err
	}

	sna := &net.IPNet{
		IP:   net.ParseIP(address),
		Mask: sn.Mask,
	}

	_, ml := sn.Mask.Size()
	a := &net.IPNet{
		IP:   sna.IP,
		Mask: net.CIDRMask(ml, ml),
	}

	return sn, sna, a, nil
}

func (d *Driver) RequestAddress(r *ipam.RequestAddressRequest) (*ipam.RequestAddressResponse, error) {
	d.log.WithField("r", r).Debug("RequestAddress()")

	subnet, addrInSubnet, addrOnly, err := getAddresses(r.Address, r.PoolID)
	if err != nil {
		return nil, err
	}

	if r.Options["RequestAddressType"] == "com.docker.network.gateway" && r.Address != "" {
		return &ipam.RequestAddressResponse{
			Address: addrInSubnet.String(),
		}, nil
	}

	// keep looking for a random address until one is found
	routes := []netlink.Route{{}}
	for len(routes) > 0 {
		if r.Address == "" {
			addrOnly.IP = iputil.RandAddr(subnet)
		}
		routes, err = netlink.RouteListFiltered(0, &netlink.Route{Dst: addrOnly}, netlink.RT_FILTER_DST)
		if err != nil {
			d.log.WithError(err).Error("failed to get routes")
			return nil, err
		}
	}
	addrInSubnet.IP = addrOnly.IP

	nr, err := d.vxrNet.GetNetworkResourceBySubnet(r.PoolID)
	if nr == nil {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get network from pool")
	}

	_, err = d.vxrNet.ConnectHost(nr.ID)
	if err != nil {
		return nil, err
	}

	gw, err := d.vxrNet.GetGatewayBySubnet(r.PoolID)
	if err != nil {
		return nil, err
	}

	log.WithField("addronly", addrOnly.String()).Debug("adding route to")
	err = netlink.RouteAdd(&netlink.Route{
		Dst: addrOnly,
		Gw:  gw.IP,
	})
	if err != nil {
		d.log.WithError(err).Error("failed to add route")
		return nil, err
	}

	return &ipam.RequestAddressResponse{
		Address: addrInSubnet.String(),
	}, nil
}

func (d *Driver) ReleaseAddress(r *ipam.ReleaseAddressRequest) error {
	d.log.WithField("r", r).Debug("ReleaseAddress()")

	_, _, addrOnly, err := getAddresses(r.Address, r.PoolID)

	gw, err := d.vxrNet.GetGatewayBySubnet(r.PoolID)
	if err != nil {
		return err
	}

	err = netlink.RouteDel(&netlink.Route{
		Dst: addrOnly,
		Gw:  gw.IP,
	})
	if err != nil {
		d.log.WithError(err).Error("failed to add route")
		return err
	}

	return nil
}
