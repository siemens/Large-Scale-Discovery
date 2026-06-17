/*
* Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
*
* Copyright (c) Siemens AG, 2016-2024.
*
* This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
* directory or visit <https://opensource.org/licenses/MIT>.
*
 */

package utils

import (
	"crypto/tls"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"math"
	"net"
	"strings"
	"syscall"

	"golang.org/x/net/nettest"
)

const NetworkSizeSkip = 327168 // Networks larger than this will be dropped
const NetworkSizeSplit = 2048  // Networks larger than this will be split into smaller subnets

// CountIpsInInput calculates the amount of possible IP addresses within a network range
func CountIpsInInput(subnet string) (uint, error) {

	// Sanitize input
	subnet = strings.TrimSpace(subnet)

	// Return 0 if subnet is empty string
	if len(subnet) == 0 {
		return 0, nil
	}

	// Return 1 if subnet actually is a single address
	if !strings.Contains(subnet, "/") || strings.HasSuffix(subnet, "/32") {
		return 1, nil
	}

	// Convert to IPNet struct
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return 0, err
	}

	// Convert mask and address from BigEndian to uint32
	first := binary.BigEndian.Uint32(ipNet.IP)
	mask := binary.BigEndian.Uint32(ipNet.Mask)

	// Calculate last address
	last := (first & mask) | (mask ^ 0xffffffff)

	// Return calculated amount
	return uint(last - first + 1), nil // Add broadcast IP to count
}

// SplitNetworkIpV4 splits a larger network range into smaller subnets of given size.
// Returns original input if it is already smaller than the given target size.
func SplitNetworkIpV4(network string, targetSize uint32) ([]string, error) {

	// Prepare subnet masks lookup
	cidrLookup := map[uint32]string{
		1:    "/32",
		2:    "/31",
		4:    "/30",
		8:    "/29",
		16:   "/28",
		32:   "/27",
		64:   "/26",
		128:  "/25",
		256:  "/24",
		512:  "/23",
		1024: "/22",
		2048: "/21",
		4096: "/20",
		8192: "/19",
		// ... bigger ones might not make sense, but can be added
	}
	if _, ok := cidrLookup[targetSize]; !ok {
		return nil, fmt.Errorf("invalid target size")
	}

	// Convert string to IPNet struct
	_, ipv4Net, err := net.ParseCIDR(network)
	if err != nil {
		return nil, err
	}

	// Calculate network size
	ones, bits := ipv4Net.Mask.Size()
	networkSize := uint32(math.Pow(2, float64(bits-ones)))

	// Return network if it is already small enough
	if targetSize >= networkSize {
		return []string{network}, nil
	}

	// Convert IPNet struct mask and address to uint32 network is BigEndian
	mask := binary.BigEndian.Uint32(ipv4Net.Mask)
	start := binary.BigEndian.Uint32(ipv4Net.IP)

	// Find the final address
	finish := (start & mask) | (mask ^ 0xffffffff)

	// Prepare memory for subnets
	var subnets []string

	// Loop through addresses as uint32
	for i := start; i <= finish; i += targetSize {

		// Convert back to net.IP
		ip := make(net.IP, 4)
		binary.BigEndian.PutUint32(ip, i)

		// Add to subnets
		subnets = append(subnets, ip.String()+cidrLookup[targetSize])
	}

	// Return result
	return subnets, nil
}

// GetOutboundIP gets preferred outbound ip of this machine by initializing a logical (fake) connection
// and reading the local address from it. By using UDP, the sample target does not actually need to exist,
// because no TCP handshake is required. Also the port does not matter.
func GetOutboundIP() string {

	// Establish logical connection, target does not actually need to be real.
	// However, this fails if the specified network is unreachable, which can
	// be the case if no default route is set (e.g., isolated systems)
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}

	// Cleanup connection
	defer func() { _ = conn.Close() }()

	// Extract local outbound IP from connection
	localAddr := conn.LocalAddr().(*net.UDPAddr)

	// Return IP address
	return localAddr.IP.String()
}

// GetLocalIp retrieves the IP address of the local default interface
func GetLocalIp() (string, error) {

	// Get routed interfaces
	routedInterface, errInterfaces := nettest.RoutedInterface("ip", net.FlagUp|net.FlagBroadcast)
	if errInterfaces != nil {
		return "", errInterfaces
	}

	// Get address of routed interface
	ifAddrs, errAddr := routedInterface.Addrs()
	if errAddr != nil {
		return "", errAddr
	}

	// Check if ip got discovered
	if len(ifAddrs) < 1 {
		return "", fmt.Errorf("no ip address found")
	}

	// Take first interface address
	ifAddr := ifAddrs[0]

	// Transform interface address (ip+netmask) into IP address
	var ip string
	switch addr := ifAddr.(type) {
	case *net.IPNet:
		ip = addr.IP.String()
	default:
		return "", fmt.Errorf("unexpected interface address")
	}

	// Return local ip
	return ip, nil
}

// SslSocket initializes an SSL socket listening for RPC connections
func SslSocket(listenAddress string, certFile string, keyFile string) (net.Listener, error) {

	// Load key files for RPC encryption
	cert, errLoad := tls.LoadX509KeyPair(certFile, keyFile)
	if errLoad != nil {
		return nil, fmt.Errorf("could not load RPC keys: %s", errLoad)
	}

	// Create the TLS conf
	tlsConf := TlsConfigFactory()

	// Set the SSL certificate
	tlsConf.Certificates = []tls.Certificate{cert}

	// Open network socket
	socket, errListen := tls.Listen("tcp", listenAddress, tlsConf)
	if errListen != nil {
		return nil, fmt.Errorf("could not open local port: %s", errListen)
	}

	// Return socket
	return socket, nil
}

// IsConnectionError detects whether a given error is one of the many types and sources of connectivity errors
func IsConnectionError(err error) bool {

	// Check if socket timeout error
	if netError, ok := err.(net.Error); ok && netError.Timeout() {
		return true
	}

	// Check if other connection error
	switch t := err.(type) {
	case *net.OpError:
		if t.Op == "dial" {
			return true
		} else if t.Op == "read" {
			return true
		}
	case syscall.Errno:
		if errors.Is(t, syscall.ECONNREFUSED) {
			return true
		}
	}

	// Return false as it seems to be a different kind of error
	return false
}
