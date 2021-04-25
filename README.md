# ServerStat
Go implementation to gather server statistics

ServerStat will pull network, interface, disk, memory, and load information and write the output to a json file. This code should run on most \*nix operating systems. Windows users may experience issues when collecting proccess information. 

#### Download and Build
```bash
go get github.com/af001/ServerStat
cd ServerStat
go build -ldflags '-w -extldflags "-static"' .
```
#### Example Output (-debug)
```bash
> ./ServerStat -debug
Hostname:  TestServ.local
Secure Machine Id: bddef3504f8eca38539c6266a8c1c25c436d71c041f49683341cdc7614e543cd
Uptime: 93h57m52.204751s
Current IP address :  192.168.0.5
Current Interface name :  en0

[+] CPU Info
Cpu User: 2.88%
Cpu System: 8.65%
Cpu Idle: 88.47%

[+] Memory Info
Memory Total: 17 GB
Memory Used: 10 GB
Memory Cached: 4.6 GB
Memory Free: 2.2 GB

[+] Load Average
Load Average 1 min: 1.94
Load Average 5 mins: 2.39
Load Average 15 mins: 2.49

[+] Interface info
Interface: 2 Name: en0 Rx: 3.0 GB bytes Tx: 188 MB bytes
Interface: 13 Name: en5 Rx: 18 kB bytes Tx: 18 kB bytes

[+] Disk info
Filesystem        Size    Used   Avail Use% Mounted on
apfs            500 GB   15 GB  263 GB   5% /
devfs           196 kB  196 kB     0 B 100% /dev
apfs            500 GB  3.2 GB  263 GB   1% /System/Volumes/VM
apfs            500 GB  459 MB  263 GB   0% /System/Volumes/Preboot
apfs            500 GB  2.2 MB  263 GB   0% /System/Volumes/Update
apfs            500 GB  217 GB  263 GB  45% /System/Volumes/Data

[+] Process info
PID     PPID    Executable
18692   8495    ServerStat
18647   1       mdworker_shared
15474   1       PerfPowerService
[...]
```
#### Example Output (/tmp/stats)
```json
> cat /tmp/stats 
{
 "MachineId": "bddef3504f8eca38539c6266a8c1c25c436d71c041f49683341cdc7614e543cd",
 "Hostname": "TestServ.local",
 "IpAddress": "192.168.0.5",
 "Device": "en0",
 "MacAddress": "8a:84:90:ae:86:e1",
 "Uptime": "93h57m52.204751s",
 "CPU": {
  "User": 2.882205513784461,
  "System": 8.646616541353383,
  "Idle": 88.47117794486216
 },
 "Memory": {
  "Total": 17177837568,
  "Used": 10375876608,
  "Cached": 4627365888,
  "Free": 2174595072
 },
 "LoadAverage": {
  "Load1": 1.93603515625,
  "Load5": 2.39111328125,
  "Load15": 2.4912109375
 },
 "Network": [
  {
   "Name": "en0",
   "Number": 2,
   "IP": "",
   "RxBytes": 2959504355,
   "TxBytes": 187581837
  },
  {
   "Name": "en5",
   "Number": 13,
   "IP": "",
   "RxBytes": 17830,
   "TxBytes": 18448
  }
 ],
 "Filesystem": [
  {
   "Name": "apfs",
   "MountedOn": "/",
   "Size": 499963170816,
   "Used": 15052201984,
   "Available": 0,
   "Free": 263137910784,
   "PercentUsed": 5.410760948414067
  }
 ],
  "Process": [
  {
   "PID": 18692,
   "PPID": 8495,
   "Executable": "ServerStat"
  },
  {
   "PID": 18647,
   "PPID": 1,
   "Executable": "mdworker_shared"
  }
 ]
}

```
