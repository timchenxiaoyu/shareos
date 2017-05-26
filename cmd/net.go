/*
 * Minio Cloud Storage, (C) 2017 Minio, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"shareos/set"
)


// IPv4 addresses of local host.
var localIP4 = mustGetLocalIP4()

// mustSplitHostPort is a wrapper to net.SplitHostPort() where error is assumed to be a fatal.
func mustSplitHostPort(hostPort string) (host, port string) {
	host, port, err := net.SplitHostPort(hostPort)
	println(err, "Unable to split host port %s", hostPort)
	return host, port
}

// mustGetLocalIP4 returns IPv4 addresses of local host.  It panics on error.
func mustGetLocalIP4() (ipList set.StringSet) {
	ipList = set.NewStringSet()
	addrs, err := net.InterfaceAddrs()
	if err != nil{
		println(err, "Unable to get IP addresses of this host.")
	}
	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}

		if ip.To4() != nil {
			ipList.Add(ip.String())
		}
	}

	return ipList
}

// getHostIP4 returns IPv4 address of given host.
func getHostIP4(host string) (ipList set.StringSet, err error) {
	ipList = set.NewStringSet()
	ips, err := net.LookupIP(host)
	if err != nil {
		return ipList, err
	}

	for _, ip := range ips {
		if ip.To4() != nil {
			ipList.Add(ip.String())
		}
	}

	return ipList, err
}

// byLastOctetValue implements sort.Interface used in sorting a list
// of ip address by their last octet value in descending order.
type byLastOctetValue []net.IP

func (n byLastOctetValue) Len() int      { return len(n) }
func (n byLastOctetValue) Swap(i, j int) { n[i], n[j] = n[j], n[i] }
func (n byLastOctetValue) Less(i, j int) bool {
	// This case is needed when all ips in the list
	// have same last octets, Following just ensures that
	// 127.0.0.1 is moved to the end of the list.
	if n[i].String() == "127.0.0.1" {
		return false
	}
	if n[j].String() == "127.0.0.1" {
		return true
	}
	return []byte(n[i].To4())[3] > []byte(n[j].To4())[3]
}

// sortIPs - sort ips based on higher octects.
// The logic to sort by last octet is implemented to
// prefer CIDRs with higher octects, this in-turn skips the
// localhost/loopback address to be not preferred as the
// first ip on the list. Subsequently this list helps us print
// a user friendly message with appropriate values.
func sortIPs(ipList []string) []string {
	if len(ipList) == 1 {
		return ipList
	}

	var ipV4s []net.IP
	var nonIPs []string
	for _, ip := range ipList {
		nip := net.ParseIP(ip)
		if nip != nil {
			ipV4s = append(ipV4s, nip)
		} else {
			nonIPs = append(nonIPs, ip)
		}
	}

	sort.Sort(byLastOctetValue(ipV4s))

	var ips []string
	for _, ip := range ipV4s {
		ips = append(ips, ip.String())
	}

	return append(nonIPs, ips...)
}

func getAPIEndpoints(serverAddr string) (apiEndpoints []string) {
	host, port := mustSplitHostPort(serverAddr)

	var ipList []string
	if host == "" {
		ipList = sortIPs(localIP4.ToSlice())
	} else {
		ipList = []string{host}
	}

	scheme := httpScheme
	//if globalIsSSL {
	//	scheme = httpsScheme
	//}

	for _, ip := range ipList {
		apiEndpoints = append(apiEndpoints, fmt.Sprintf("%s://%s:%s", scheme, ip, port))
	}

	return apiEndpoints
}

// checkPortAvailability - check if given port is already in use.
// Note: The check method tries to listen on given port and closes it.
// It is possible to have a disconnected client in this tiny window of time.
func checkPortAvailability(port string) (err error) {
	// Return true if err is "address already in use" error.
	isAddrInUseErr := func(err error) (b bool) {
		if opErr, ok := err.(*net.OpError); ok {
			if sysErr, ok := opErr.Err.(*os.SyscallError); ok {
				if errno, ok := sysErr.Err.(syscall.Errno); ok {
					b = (errno == syscall.EADDRINUSE)
				}
			}
		}

		return b
	}

	network := []string{"tcp", "tcp4", "tcp6"}
	for _, n := range network {
		l, err := net.Listen(n, net.JoinHostPort("", port))
		if err == nil {
			// As we are able to listen on this network, the port is not in use.
			// Close the listener and continue check other networks.
			if err = l.Close(); err != nil {
				return err
			}
		} else if isAddrInUseErr(err) {
			// As we got EADDRINUSE error, the port is in use by other process.
			// Return the error.
			return err
		}
	}

	return nil
}

// extractHostPort - extracts host/port from many address formats
// such as, ":9000", "localhost:9000", "http://localhost:9000/"
func extractHostPort(hostAddr string) (string, string, error) {
	var addr, scheme string

	if hostAddr == "" {
		return "", "", errors.New("unable to process empty address")
	}

	// Parse address to extract host and scheme field
	u, err := url.Parse(hostAddr)
	if err != nil {
		// Ignore scheme not present error
		if !strings.Contains(err.Error(), "missing protocol scheme") {
			return "", "", err
		}
	} else {
		addr = u.Host
		scheme = u.Scheme
	}

	// Use the given parameter again if url.Parse()
	// didn't return any useful result.
	if addr == "" {
		addr = hostAddr
		scheme = "http"
	}

	// At this point, addr can be one of the following form:
	//	":9000"
	//	"localhost:9000"
	//	"localhost" <- in this case, we check for scheme

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		if !strings.Contains(err.Error(), "missing port in address") {
			return "", "", err
		}

		host = addr

		switch scheme {
		case "https":
			port = "443"
		case "http":
			port = "80"
		default:
			return "", "", errors.New("unable to guess port from scheme")
		}
	}

	return host, port, nil
}

// isLocalHost - checks if the given parameter
// correspond to one of the local IP of the
// current machine
func isLocalHost(host string) (bool, error) {
	hostIPs, err := getHostIP4(host)
	if err != nil {
		return false, err
	}

	// If intersection of two IP sets is not empty, then the host is local host.
	isLocal := !localIP4.Intersection(hostIPs).IsEmpty()
	return isLocal, nil
}

// sameLocalAddrs - returns true if two addresses, even with different
// formats, point to the same machine, e.g:
//   ':9000' and 'http://localhost:9000/' will return true
func sameLocalAddrs(addr1, addr2 string) (bool, error) {

	// Extract host & port from given parameters
	host1, port1, err := extractHostPort(addr1)
	if err != nil {
		return false, err
	}
	host2, port2, err := extractHostPort(addr2)
	if err != nil {
		return false, err
	}

	var addr1Local, addr2Local bool

	if host1 == "" {
		// If empty host means it is localhost
		addr1Local = true
	} else {
		// Host not empty, check if it is local
		if addr1Local, err = isLocalHost(host1); err != nil {
			return false, err
		}
	}

	if host2 == "" {
		// If empty host means it is localhost
		addr2Local = true
	} else {
		// Host not empty, check if it is local
		if addr2Local, err = isLocalHost(host2); err != nil {
			return false, err
		}
	}

	// If both of addresses point to the same machine, check if
	// have the same port
	if addr1Local && addr2Local {
		if port1 == port2 {
			return true, nil
		}
	}
	return false, nil
}

// CheckLocalServerAddr - checks if serverAddr is valid and local host.
func CheckLocalServerAddr(serverAddr string) error {
	host, port, err := net.SplitHostPort(serverAddr)
	if err != nil {
		return err
	}

	// Check whether port is a valid port number.
	p, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("invalid port number")
	} else if p < 1 || p > 65535 {
		return fmt.Errorf("port number must be between 1 to 65535")
	}

	if host != "" {
		isLocalHost, err := isLocalHost(host)
		if err != nil {
			return err
		}
		if !isLocalHost {
			return fmt.Errorf("host in server address should be this server")
		}
	}

	return nil
}
