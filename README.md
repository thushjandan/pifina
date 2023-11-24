![Build status](https://github.com/thushjandan/pifina/actions/workflows/build.yml/badge.svg?branch=main)
![GitHub](https://img.shields.io/github/license/thushjandan/pifina)
![GitHub release](https://img.shields.io/github/v/release/thushjandan/pifina)

# Performance Introspector for in-network applications (PIFINA)
This is a performance framework to introspect in-network applications written in [P4 programming language](https://p4.org) running on `Intel Tofino` powered switches. The framework has been developed and tested for the Intel Tofino architecture version 2, but it is backwards compatible to Tofino 1.

## Installation
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

See for more information on [pifina.app](https://pifina.app)
## Usage
1. Instrumentize your P4 application with PIFINA
```bash
# Match on P4 header fields hdr.ipv4.protocol, hdr.ipv4.dstAddr & hdr.ipv4.srcAddr
# Write the generated P4 libraries in src/myP4app/include
user@myworkstation$ pifina generate -k hdr.ipv4.protocol:exact -k hdr.ipv4.dstAddr:ternary -k hdr.ipv4.srcAddr:ternary -o src/myP4app/include
# Use command help to see the options
user@myworkstation$ pifina generate -h
```
2. Run your P4 app enriched with PIFINA on your Tofino switch
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
4. Start the tofino probe on the Tofino switch
```bash
# Start the tofino probe. The P4 app name must be given with the flag p4name
sde@tofino$ pifina-tofino-probe -p4name myP4app -server pifina-collector.local:8654
# Use command help to see the options like bfrt server address or pifina
sde@tofino$ pifina-tofino-probe -h
```
5. Optional: Start the NIC collector on your sender and receiver
  * This component depends on NVIDIA NEO-Host SDK and it must be already installed!
```bash
# List all available Mellanox ConnectX NICs
admin@server1$ pifina nic list
# Collect metrics from mlx5_1 NIC and send metrics to PIFINA collector
admin@server1$ pifina nic collect -d mlx5_1 -s pifina-collector.local:8654
# Use command help to see the options
admin@server1$ pifina nic -h
```

Check the user manual on [pifina.app](https://pifina.app)

## Authors and acknowledgment
* Thushjandan Ponnudurai
* Alberto Lerner

### Acknowledgments
* Skiplist implementation from github.com/sean-public/fast-skiplist

## Contributing

Pull requests are welcome. For major changes, please open an issue first
to discuss what you would like to change.

## License

[MIT](https://choosealicense.com/licenses/mit/)
