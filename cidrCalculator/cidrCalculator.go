package cidrCalculator

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/apparentlymart/go-cidr/cidr"
	"net"
	"sort"
	"strconv"
	"strings"
)

type cidrCalculator struct {
	currentSubnets       *map[string]string
	prefixLength         int8
	baseCidrRange        string
	baseCidrPrefixLength int8
	baseIPNet            *net.IPNet
}

func New(currentSubnets *map[string]string, prefixLength int8, baseCidrRange string) (cidrCalculator, error) {
	baseCidrPrefixLength, err := strconv.ParseInt(strings.Split(baseCidrRange, "/")[1], 10, 8)
	if err != nil {
		return cidrCalculator{}, err
	}
	_, baseIPNet, _ := net.ParseCIDR(baseCidrRange)
	return cidrCalculator{currentSubnets, prefixLength, baseCidrRange, int8(baseCidrPrefixLength) - 1, baseIPNet}, nil
}

//TODO: implement!
func (c cidrCalculator) GetNextNetmask() (string, error) {
	if c.prefixLength > 32 || c.prefixLength < 0 {
		return "", errors.New("prefixLength must be an integer between 0 and 32")
	}
	ipNets := make([]*net.IPNet, 0, len(*c.currentSubnets))
	for _, value := range *c.currentSubnets {
		_, ipNet, err := net.ParseCIDR(value)
		if err != nil {
			return "", err
		}
		ipNets = append(ipNets, ipNet)
	}
	sort.Slice(ipNets, func(i, j int) bool {
		return bytes.Compare(ipNets[i].IP, ipNets[j].IP) > 0
	})
	//ipNetsCidr := make([]string, 0, len(*currentSubnets))
	//for _, ipNet := range ipNets {
	//	ipNetsCidr = append(ipNetsCidr, ipNet.String())
	//}
	var nextIPNet *net.IPNet
	firstCidrSubnet, err := cidr.Subnet(c.baseIPNet, int(c.prefixLength-c.baseCidrPrefixLength-1), 0)
	if err != nil {
		return "", err
	}
	firstCidrSubnetFirstIP, firstCidrSubnetLastIP := cidr.AddressRange(firstCidrSubnet)
	if len(ipNets) == 0 || (!ipNets[len(ipNets)-1].Contains(firstCidrSubnetFirstIP) && !ipNets[len(ipNets)-1].Contains(firstCidrSubnetLastIP)) {
		nextIPNet = firstCidrSubnet
	} else {
		nextIPNet, err = c.recursivelyFindNextNetmask(&ipNets, c.prefixLength)
	}
	if err != nil {
		return "", err
	}
	//nextNetBaseIP, _, _ := net.ParseCIDR(nextNetmask)
	if !c.baseIPNet.Contains(nextIPNet.IP) {
		return "", fmt.Errorf("baseCidrRange %s is exhausted!", c.baseCidrRange)
	}
	return nextIPNet.String(), nil
}

func (c cidrCalculator) recursivelyFindNextNetmask(ipNets *[]*net.IPNet, searchPrefixLength int8) (*net.IPNet, error) {
	if searchPrefixLength <= c.baseCidrPrefixLength {
		lastSubnet := (*ipNets)[len(*ipNets)-1]
		nextIPNet, exhausted := cidr.NextSubnet(lastSubnet, int(c.prefixLength))
		if exhausted {
			return nil, fmt.Errorf("Maximum IP exhausted!")
		}
		return nextIPNet, nil
	}
	_, maskNet, _ := net.ParseCIDR(fmt.Sprintf("0.0.0.0/%d", searchPrefixLength))
	mask := maskNet.Mask
	for index, ipNet := range *ipNets {
		if bytes.Compare(mask, ipNet.Mask) <= 0 {
			//if index < len(*ipNets)-1 && bytes.Equal(mask, (*ipNets)[index+1].Mask) {
			//	continue
			//}
			nextIPNet, exhausted := cidr.NextSubnet(ipNet, int(c.prefixLength))
			if exhausted {
				return nil, fmt.Errorf("Maximum IP exhausted!")
			}
			if index < len(*ipNets)-1 {
				if !nextIPNet.Contains((*ipNets)[index+1].IP) {
					return nextIPNet, nil
				}
			} else {
				return nextIPNet, nil
			}
			return c.recursivelyFindNextNetmask(ipNets, searchPrefixLength-1)
		}
	}
	return c.recursivelyFindNextNetmask(ipNets, searchPrefixLength-1)
}
