package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/denisbrodbeck/machineid"
	human "github.com/dustin/go-humanize"
	"github.com/mackerelio/go-osstat/cpu"
	"github.com/mackerelio/go-osstat/loadavg"
	"github.com/mackerelio/go-osstat/memory"
	"github.com/mackerelio/go-osstat/network"
	"github.com/mackerelio/go-osstat/uptime"
	"github.com/mitchellh/go-ps"
	"github.com/shirou/gopsutil/disk"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"time"
)

var (
	logfile 	= flag.String("f", "/tmp/stats", "Location and name of the logfile")
	debug		= flag.Bool("debug", false, "Enable debugging mode")
	verbose		= flag.Bool("verbose", false, "Enable verbose debugging")
)

type Survey struct {
	MachineId		string
	Hostname 		string
	IpAddress		string
	Device   		string
	MacAddress		string
	Uptime 			string
	CPU				CPU
	Memory 			Memory
	LoadAverage 	LoadAverage
	Network 		[]Interface
	Filesystem 		[]Filesystem
	Process 		[]Process
}

type CPU struct {
	User, System, Idle float64
}

type Memory struct {
	Total, Used, Cached, Free uint64
}

type LoadAverage struct {
	Load1, Load5, Load15 float64
}

type Interface struct {
	Name string
	Number	int
	IP   	string
	RxBytes, TxBytes uint64
}

type Filesystem struct {
	Name, MountedOn string
	Size, Used, Available, Free uint64
	PercentUsed	float64
}

type Process struct {
	PID int
	PPID int
	Executable string
}

func init() {
	flag.Usage = func() {
		fmt.Println("Usage: ./ss [ ... ]\n\nParameters:")
		flag.PrintDefaults()
	}
}

func GetHardwareMac(ip string) (string,string,error) {
	var currentNetworkHardwareName string

	interfaces, _ := net.Interfaces()
	for _, interf := range interfaces {

		if addrs, err := interf.Addrs(); err == nil {
			for index, addr := range addrs {
				if *debug && *verbose {fmt.Println("[", index, "]", interf.Name, ">", addr)}

				if strings.Contains(addr.String(), ip) {
					if *debug {fmt.Println("Current Interface name : ", interf.Name)}
					currentNetworkHardwareName = interf.Name
				}
			}
		}
	}

	netInterface, err := net.InterfaceByName(currentNetworkHardwareName)

	if err != nil {os.Exit(1)}

	name := netInterface.Name
	macAddress := netInterface.HardwareAddr

	// verify if the MAC address can be parsed properly
	hwAddr, err := net.ParseMAC(macAddress.String())
	if err != nil {os.Exit(1)}

	return name, hwAddr.String(), nil
}

func GetCurrnentIp() (string,error) {
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		fmt.Println(err)
	}

	var currentIP string

	for _, address := range addrs {

		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				if *debug {
					fmt.Println("Current IP address : ", ipnet.IP.String())
				}
				currentIP = ipnet.IP.String()
			}
		}
	}
	return currentIP,nil
}

func main() {
	flag.Parse()

	survey := Survey{}

	// Get Hostname
	name, err := os.Hostname()
	if err != nil {os.Exit(1)}
	survey.Hostname = name
	if *debug {fmt.Println("Hostname: ", name)}

	// Get secure machine id
	id, err := machineid.ProtectedID("serverstats")
	if err != nil {os.Exit(1)}
	survey.MachineId = id
	if *debug {fmt.Printf("Secure Machine Id: %s\n", id)}

	// Get server uptime
	upstat, err := uptime.Get()
	if err != nil {os.Exit(1)}
	survey.Uptime = upstat.String()
	if *debug {fmt.Printf("Uptime: %s\n", upstat.String())}

	// Get IP and Mac address
	ip, err := GetCurrnentIp()
	survey.IpAddress = ip
	if err != nil {os.Exit(1)}
	device, mac, err := GetHardwareMac(ip)
	if err != nil {os.Exit(1)}
	survey.MacAddress = mac
	survey.Device = device

	// Get CPU Info
	before, err := cpu.Get()
	if err != nil {os.Exit(1)}

	time.Sleep(time.Duration(1) * time.Second)
	after, err := cpu.Get()
	if err != nil {os.Exit(1)}
	total := float64(after.Total - before.Total)

	c := CPU{}
	c.User = float64(after.User-before.User)/total*100
	c.System = float64(after.System-before.System)/total*100
	c.Idle = float64(after.Idle-before.Idle)/total*100
	survey.CPU = c

	if *debug {
		fmt.Println("\n[+] CPU Info")
		fmt.Printf("Cpu User: %.2f%%\n", float64(after.User-before.User)/total*100)
		fmt.Printf("Cpu System: %.2f%%\n", float64(after.System-before.System)/total*100)
		fmt.Printf("Cpu Idle: %.2f%%\n", float64(after.Idle-before.Idle)/total*100)
	}

	// Get Memory Info
	memstats, err := memory.Get()
	if err != nil {os.Exit(1)}

	m := Memory{}
	m.Total = memstats.Total
	m.Used = memstats.Used
	m.Cached = memstats.Cached
	m.Free = memstats.Free
	survey.Memory = m

	if *debug {
		fmt.Println("\n[+] Memory Info")
		fmt.Printf("Memory Total: %s\n", human.Bytes(memstats.Total))
		fmt.Printf("Memory Used: %s\n", human.Bytes(memstats.Used))
		fmt.Printf("Memory Cached: %s\n", human.Bytes(memstats.Cached))
		fmt.Printf("Memory Free: %s\n", human.Bytes(memstats.Free))
	}

	ldavg, err := loadavg.Get()
	if err != nil {os.Exit(1)}

	l := LoadAverage{}
	l.Load1 = ldavg.Loadavg1
	l.Load5 = ldavg.Loadavg5
	l.Load15 = ldavg.Loadavg15
	survey.LoadAverage = l

	if *debug {
		fmt.Println("\n[+] Load Average")
		fmt.Printf("Load Average 1 min: %.2f\n", ldavg.Loadavg1)
		fmt.Printf("Load Average 5 mins: %.2f\n", ldavg.Loadavg5)
		fmt.Printf("Load Average 15 mins: %.2f\n", ldavg.Loadavg15)
	}

	// Get Interface info
	netstat, err := network.Get()
	if err != nil {os.Exit(1)}

	var x []Interface
	fmt.Println("\n[+] Interface info")
	for n, s := range netstat {
		if s.TxBytes > 0 && s.RxBytes > 0 {
			i := Interface{}
			i.Number = n
			i.Name = s.Name
			i.TxBytes = s.TxBytes
			i.RxBytes = s.RxBytes
			x = append(x, i)

			if *debug {
				fmt.Printf("Interface: %d Name: %s Rx: %s bytes Tx: %s bytes\n", n,
					s.Name, human.Bytes(s.RxBytes), human.Bytes(s.TxBytes))
			}
		}
	}
	survey.Network = x

	// Get disk info
	formatter := "%-14s %7s %7s %7s %4s %s\n"
	if *debug {
		fmt.Println("\n[+] Disk info")
		fmt.Printf(formatter, "Filesystem", "Size", "Used", "Avail", "Use%", "Mounted on")
	}

	parts, err := disk.Partitions(true)
	if err != nil {os.Exit(1)}

	var y []Filesystem
	for _, p := range parts {
		device := p.Mountpoint
		s, _ := disk.Usage(device)

		if s.Total == 0 {
			continue
		}

		percent := fmt.Sprintf("%2.f%%", s.UsedPercent)

		d := Filesystem{}
		d.Name = s.Fstype
		d.Size = s.Total
		d.Used = s.Used
		d.Free = s.Free
		d.PercentUsed = s.UsedPercent
		d.MountedOn = p.Mountpoint
		y = append(y, d)

		if *debug {
			fmt.Printf(formatter,
				s.Fstype,
				human.Bytes(s.Total),
				human.Bytes(s.Used),
				human.Bytes(s.Free),
				percent,
				p.Mountpoint,
			)
		}
	}
	survey.Filesystem = y

	// Get running processes
	processList, err := ps.Processes()
	if err != nil {os.Exit(1)}

	if *debug {
		fmt.Println("\n[+] Process info")
		fmt.Println("PID\tPPID\tExecutable")
	}

	var z []Process
	for x := range processList {
		var process ps.Process
		process = processList[x]
		if *debug {fmt.Printf("%d\t%d\t%s\n",process.Pid(), process.PPid(), process.Executable())}

		p := Process{}
		p.PID = process.Pid()
		p.PPID = process.PPid()
		p.Executable = process.Executable()
		z = append(z, p)
	}
	survey.Process = z

	file, _ := json.MarshalIndent(survey, "", " ")
	_ = ioutil.WriteFile(*logfile, file, 0644) 
}

