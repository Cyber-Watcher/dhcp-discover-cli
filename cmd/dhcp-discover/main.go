package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"time"
)

const (
	dhcpServerPort = 67
	dhcpClientPort = 68
	minPacketSize  = 300
)

var (
	timeoutFlag     = flag.Duration("timeout", 8*time.Second, "listen timeout (e.g. 4s, 1m)")
	ifaceNameFlag   = flag.String("iface", "", "network interface name (exact)")
	ifaceIndexFlag  = flag.Int("iface-index", -1, "network interface index (from --show-interfaces)")
	showIfacesFlag  = flag.Bool("show-interfaces", false, "list available network interfaces and exit")
	verboseFlag     = flag.Bool("verbose", false, "enable verbose logging (stdout + dhcp-discover.log)")
	retryCount      = flag.Int("retry", 3, "number of discovery attempts")
)

func init() {
	flag.BoolVar(showIfacesFlag, "si", false, "alias for --show-interfaces")
	flag.BoolVar(verboseFlag, "v", false, "alias for --verbose")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), `Usage: %s [options]
    
Options:
  --show-interfaces, -si      list available network interfaces and exit
  --iface name                select interface by exact name
  --iface-index n             select interface by number from show-interfaces (1-based)
  --timeout duration          listen timeout, default 8s
  --retry n                   number of discovery attempts, default 3
  --verbose, -v               enable verbose logging to console and dhcp-discover.log
  --help, -h                  show this help
`, os.Args[0])
	}
}

func main() {
	flag.Parse()
	rand.Seed(time.Now().UnixNano())

	// Verbose logging
	if *verboseFlag {
		log.SetFlags(log.LstdFlags | log.Lmicroseconds)
		
		// Creating a multi-writer for console and file output
		writers := []io.Writer{os.Stdout}
		
		exe, _ := os.Executable()
		logf := filepath.Join(filepath.Dir(exe), "dhcp-discover.log")
		if f, err := os.OpenFile(logf, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644); err == nil {
			writers = append(writers, f)
			defer f.Close()
		} else {
			log.Printf("Failed to open log file: %v", err)
		}
		
		multiWriter := io.MultiWriter(writers...)
		log.SetOutput(multiWriter)
		
		log.Println("=== Start DHCP Discover ===")
	}

	// Get all the active interfaces
	activeIfaces := getActiveInterfaces()

	// Показать интерфейсы с IP-адресами
	if *showIfacesFlag {
		if len(activeIfaces) == 0 {
			fmt.Println("No active network interfaces found")
			return
		}
		
		fmt.Println("Available network interfaces:")
		for i, ifc := range activeIfaces {
			addrs, _ := ifc.Addrs()
			var ips []string
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok {
					ips = append(ips, ipnet.IP.String())
				}
			}
			
			fmt.Printf("  %d) %s\n", i+1, ifc.Name)
			for _, ip := range ips {
				fmt.Printf("      %s\n", ip)
			}
		}
		return
	}

	if len(activeIfaces) == 0 {
		log.Fatal("no active network interfaces found")
	}

	// Choosing the interface
	iface, err := selectInterface(activeIfaces)
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("Using interface: %s (MAC: %s)\n", iface.Name, iface.HardwareAddr)
	if *verboseFlag {
		log.Printf("Selected interface: %s (MAC: %s)", iface.Name, iface.HardwareAddr)
	}

	// Generating a unique XID
	xid := rand.Uint32()
	servers := make(map[string]struct{})

	for attempt := 0; attempt < *retryCount; attempt++ {
		if *verboseFlag {
			log.Printf("Attempt %d/%d", attempt+1, *retryCount)
		}

		// Creating a socket
		conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: dhcpClientPort})
		if err != nil {
			log.Printf("socket error: %v", err)
			continue
		}

		// Always use global broadcast
		bc := net.IPv4bcast
		if *verboseFlag {
			log.Printf("Using broadcast: %s", bc)
		}

		// Building a DHCPDISCOVER
		packet := buildDiscoverPacket(iface, xid)

		// Sending
		if _, err := conn.WriteToUDP(packet, &net.UDPAddr{IP: bc, Port: dhcpServerPort}); err != nil {
			log.Printf("send error: %v", err)
			conn.Close()
			continue
		}

		// Reading the answers
		conn.SetReadDeadline(time.Now().Add(*timeoutFlag / time.Duration(*retryCount)))
		found := readResponses(conn, xid, servers, *verboseFlag)
		conn.Close()
		
		if *verboseFlag {
			log.Printf("Attempt %d: found %d servers", attempt+1, found)
		}
		
		time.Sleep(500 * time.Millisecond)
	}

	// Output of results
	if len(servers) == 0 {
		fmt.Println("no DHCP servers found")
	} else {
		fmt.Println("DHCP servers:")
		for ip := range servers {
			fmt.Printf("  %s\n", ip)
		}
	}
	if *verboseFlag {
		log.Println("=== End DHCP Discover ===")
	}
}

func getActiveInterfaces() []net.Interface {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}

	var active []net.Interface
	for _, ifc := range ifaces {
		if ifc.Flags&net.FlagUp != 0 && ifc.Flags&net.FlagLoopback == 0 {
			active = append(active, ifc)
		}
	}
	return active
}

func selectInterface(activeIfaces []net.Interface) (net.Interface, error) {
	if *ifaceIndexFlag > 0 {
		idx := *ifaceIndexFlag - 1
		if idx < 0 || idx >= len(activeIfaces) {
			return net.Interface{}, fmt.Errorf("iface-index %d out of range (1-%d)", 
				*ifaceIndexFlag, len(activeIfaces))
		}
		return activeIfaces[idx], nil
	}

	if *ifaceNameFlag != "" {
		for _, ifc := range activeIfaces {
			if ifc.Name == *ifaceNameFlag {
				return ifc, nil
			}
		}
		return net.Interface{}, fmt.Errorf("interface named %q not found", *ifaceNameFlag)
	}

	if len(activeIfaces) > 0 {
		return activeIfaces[0], nil
	}
	
	return net.Interface{}, fmt.Errorf("no active interfaces available")
}

func buildDiscoverPacket(iface net.Interface, xid uint32) []byte {
	packet := make([]byte, minPacketSize)
	
	// BOOTP header
	packet[0] = 1  // OP: BOOTREQUEST
	packet[1] = 1  // HTYPE: Ethernet
	packet[2] = 6  // HLEN: MAC length
	packet[3] = 0  // HOPS
	
	binary.BigEndian.PutUint32(packet[4:8], xid) // XID
	
	// Always set the broadcast flag.
	binary.BigEndian.PutUint16(packet[10:12], 1<<15)
	
	// Client MAC address
	if len(iface.HardwareAddr) >= 6 {
		copy(packet[28:34], iface.HardwareAddr[:6])
	}
	
	// Magic cookie
	copy(packet[236:240], []byte{99, 130, 83, 99})
	
	// DHCP options
	options := []byte{
		53, 1, 1,    // DHCP Discover
		55, 4,        // Parameter request list
		1, 3, 6, 15,  // Subnet, Router, DNS, Domain
		255,            // End option
	}
	copy(packet[240:], options)
	
	return packet
}

func readResponses(conn *net.UDPConn, xid uint32, servers map[string]struct{}, verbose bool) int {
	found := 0
	buf := make([]byte, 1500)
	
	for {
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				break
			}
			if verbose {
				log.Printf("read error: %v", err)
			}
			continue
		}
		
		if verbose {
			log.Printf("Received %d bytes from %s", n, addr.IP)
		}
		
		// Checking the minimum length
		if n < 240 {
			if verbose {
				log.Printf("Packet too small: %d < 240", n)
			}
			continue
		}
		
		// Checking the magic cookie
		if buf[236] != 99 || buf[237] != 130 || buf[238] != 83 || buf[239] != 99 {
			if verbose {
				log.Printf("Invalid magic cookie: %v", buf[236:240])
			}
			continue
		}
		
		// XID verification
		if binary.BigEndian.Uint32(buf[4:8]) != xid {
			if verbose {
				log.Printf("XID mismatch: expected %d, got %d", 
					xid, binary.BigEndian.Uint32(buf[4:8]))
			}
			continue
		}
		
		// Search for an option 53 (DHCP Message Type)
		msgType := byte(0)
		for i := 240; i < n; {
			if buf[i] == 255 { // End option
				break
			}
			
			optType := buf[i]
			if i+1 >= n {
				break
			}
			
			optLen := int(buf[i+1])
			if i+2+optLen > n {
				break
			}
			
			if optType == 53 && optLen >= 1 { // DHCP Message Type
				msgType = buf[i+2]
				break
			}
			
			i += optLen + 2
		}
		
		if msgType != 2 { // DHCPOFFER = 2
			if verbose {
				log.Printf("Not DHCPOFFER: message type %d", msgType)
			}
			continue
		}
		
		// Adding a server
		serverIP := addr.IP.String()
		if _, exists := servers[serverIP]; !exists {
			servers[serverIP] = struct{}{}
			found++
			if verbose {
				log.Printf("Found DHCP server: %s", serverIP)
			}
		}
	}
	return found
}