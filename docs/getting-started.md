# Getting started

## Installation
### Install using precompiled binaries
1. Download compiled PIFINA binaries from [Github](https://github.com/thushjandan/pifina/releases/latest)
2. Untar the archive
```
cd /tmp
tar -xzf pifina_Linux_x86_64.tar.gz
mv pifina /usr/local/bin
mv pifina-tofino-probe /usr/local/bin
```

### Install from source
1. Install prequisites
  * Install the latest [Golang compiler](https://go.dev/doc/install)
  * Install the latest [Protobuf compiler](https://grpc.io/docs/protoc-installation/)
  * Install protobuf plugin for golang
  ```bash
  go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
  ```
  * Install the latest [Node.js LTS](https://nodejs.org/en/download)
2. Compile using makefile
```bash
cd pifina-sdk/
make build
```
Now you can find two binaries in the build folder.

## How to use PIFINA
1. Instrumentize your P4 application with PIFINA
```bash
# Match on P4 header fields hdr.ipv4.protocol, hdr.ipv4.dstAddr & hdr.ipv4.srcAddr
# Write the generated P4 libraries in src/myP4app/include
user@myworkstation$ pifina generate -k hdr.ipv4.protocol:exact -k hdr.ipv4.dstAddr:ternary -k hdr.ipv4.srcAddr:ternary -o src/myP4app/include
# Use command help to see the options
user@myworkstation$ pifina generate -h
```
Three header fields from the IPv4 header are used as keys in this example for the match action table. Ternary match type is used on the ipv4 source and destination address header fields and exact match type is used for the protocol header field. The generated P4<sub>16</sub> code is stored under src/myP4app/include
2. Run your P4 app enriched with PIFINA on your Tofino switch
```bash
~/start_switch.sh
```
3. Start the PIFINA collector on a remote server
```bash
# Quick and dirty way
# Create self signed certificate
admin@collector$ mkdir assets
admin@collector$ openssl req -x509 -newkey  ec -pkeyopt ec_paramgen_curve:prime256v1 \
	-keyout assets/key.pem -out assets/cert.pem -sha256 -days 3650 -nodes \
	-subj "/C=CH/O=Pifina/CN=PifinaServer"
admin@collector$ pifina serve
# Start collector with a signed TLS certificate
admin@collector$ pifina serve --key privatekey.pem --cert letsencrypt_cert.pem
# Use command help to see the options
admin@collector$ pifina serve -h
```
The web application is then reachable on port 8655 over https (https://pifina-collector.local:8655) and metrics are received over port 8654
4. Start the tofino probe on the Tofino switch
```bash
# Start the tofino probe. The P4 app name must be given with the flag p4name
sde@tofino$ pifina-tofino-probe -p4name myP4app -server pifina-collector.local:8654
# Use command help to see the options like bfrt server address or pifina
sde@tofino$ pifina-tofino-probe -h
```
The mandatory command line flag -p4name defines the loaded P4 application name. The flag -server defines the address and the port of the collector server.
5. Optional: Start the NIC collector on your sender and receiver
  * This component uses the NVIDIA NEO-Host SDK and that SDK must be already installed!
```bash
# List all available Mellanox ConnectX NICs
admin@server1$ pifina nic list
# Collect metrics from mlx5_1 NIC and send metrics to PIFINA collector
admin@server1$ pifina nic collect -d mlx5_1 -s pifina-collector.local:8654
# Use command help to see the options
admin@server1$ pifina nic -h
```
