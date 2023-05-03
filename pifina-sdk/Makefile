BFRT_COMPILED_PROTO=internal/dataplane/tofino/protos/bfruntime/bfruntime_grpc.pb.go
PIFINA_COMPILED_PROTO=pkg/sink/protos/pifina/pifina.pb.go

build: $(BFRT_COMPILED_PROTO) $(PIFINA_COMPILED_PROTO)
	mkdir -p builds
	go build -o builds/pifina-cp cmd/pifina-cp/main.go
	go build -o builds/pifina-cli cmd/pifina-cli/main.go

run: $(BFRT_COMPILED_PROTO) $(PIFINA_COMPILED_PROTO)
	go run cmd/pifina-cp/main.go -level debug

$(BFRT_COMPILED_PROTO):
	protoc --go-grpc_opt=Mbfruntime.proto=bfruntime/ \
	--go_opt=Mbfruntime.proto=bfruntime/ \
	--go_out=./internal/dataplane/tofino/protos \
	--go-grpc_out=./internal/dataplane/tofino/protos \
	--proto_path=/home/dev/bf-sde-9.9.0/install/share/bf_rt_shared/proto/ \
	--proto_path=/home/dev/bf-sde-9.9.0/pkgsrc/bf-drivers/third-party/ \
	/home/dev/bf-sde-9.9.0/install/share/bf_rt_shared/proto/bfruntime.proto

$(PIFINA_COMPILED_PROTO):
	protoc --go_opt=Mpifina.proto=pifina/ \
	--go_out=./pkg/sink/protos/pifina/ \
	--proto_path=./pkg/sink/protos/ \
	./pkg/sink/protos/pifina.proto