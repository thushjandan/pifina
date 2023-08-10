BFRT_COMPILED_PROTO=internal/dataplane/tofino/protos/bfruntime/bfruntime_grpc.pb.go
PIFINA_COMPILED_PROTO=pkg/model/protos/pifina/pifina.pb.go
SELF_SIGNED_TLS_CERTS=assets/cert.pem

build: $(BFRT_COMPILED_PROTO) $(PIFINA_COMPILED_PROTO) $(SELF_SIGNED_TLS_CERTS)
	mkdir -p builds
	$(MAKE) -C frontend/ all
	go mod download
	go build -o builds/pifina cmd/pifina/main.go
	go build -o builds/pifina-tofino-probe cmd/pifina-tofino/main.go

run: $(BFRT_COMPILED_PROTO) $(PIFINA_COMPILED_PROTO)
	go run cmd/pifina-tofino/main.go -level debug

$(BFRT_COMPILED_PROTO):
	protoc --go-grpc_opt=Mbfruntime.proto=bfruntime/ \
	--go_opt=Mbfruntime.proto=bfruntime/ \
	--go_out=./internal/dataplane/tofino/protos \
	--go-grpc_out=./internal/dataplane/tofino/protos \
	--proto_path=~/bf-sde-9.9.0/install/share/bf_rt_shared/proto/ \
	--proto_path=~/bf-sde-9.9.0/pkgsrc/bf-drivers/third-party/ \
	~/bf-sde-9.9.0/install/share/bf_rt_shared/proto/bfruntime.proto

$(PIFINA_COMPILED_PROTO):
	protoc --go_opt=Mpifina.proto=pifina/ \
	--go_out=./pkg/model/protos/pifina/ \
	--proto_path=./pkg/model/protos/ \
	./pkg/model/protos/pifina.proto

$(SELF_SIGNED_TLS_CERTS):
	openssl req -x509 -newkey  ec -pkeyopt ec_paramgen_curve:prime256v1 \
	-keyout assets/key.pem -out assets/cert.pem -sha256 -days 3650 -nodes \
	-subj "/C=CH/O=Pifina/CN=PifinaServer"