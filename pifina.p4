/* -*- P4_16 -*- */
#include <core.p4>
/* TOFINO Native architecture */
#include <t2na.p4>

// Size of lookup table in bits. 7 bits = 127 entries
#define PF_TABLE_SIZE_WIDTH 7
#define PF_TABLE_SIZE 1<<PF_TABLE_SIZE_WIDTH

const bit<16> TYPE_IPV4 = 0x800;

/*************************************************************************
*********************** H E A D E R S  ***********************************
*************************************************************************/

typedef bit<48> macAddr_t;
typedef bit<32> ip4Addr_t;
// jshint ignore:start
typedef bit<PF_TABLE_SIZE_WIDTH> pf_stats_width_t;

header ethernet_t {
    macAddr_t dstAddr;
    macAddr_t srcAddr;
    bit<16>   etherType;
}

header ipv4_t {
    bit<4>    version;
    bit<4>    ihl;
    bit<8>    diffserv;
    bit<16>   totalLen;
    bit<16>   identification;
    bit<3>    flags;
    bit<13>   fragOffset;
    bit<8>    ttl;
    bit<8>    protocol;
    bit<16>   hdrChecksum;
    ip4Addr_t srcAddr;
    ip4Addr_t dstAddr;
}

header tcp_t {
    bit<16> src_port;
    bit<16> dst_port;
}

header udp_t {
    bit<16> src_port;
    bit<16> dst_port;
}

header pf_control_t {
    bool pfIsMatch;
    pf_stats_width_t pfSessionId;
}

struct l4_lookup_t {
    bit<16> pfLayer4Word1; // Layer 4 destination port
    bit<16> pfLayer4Word2; // Layer 4 source port
}

struct ingress_headers_t {
    pf_control_t        pfControl;
    ethernet_t          ethernet;
    ipv4_t              ipv4;
}

struct ingress_metadata_t {
    bool pfIsMatch;
    pf_stats_width_t pfSessionId;
    bit<32> origHdrLength;
    bit<16> pfTargetProtocol;
    bit<128> pfDstAddr;
    bit<128> pfSrcAddr;
    l4_lookup_t pfL4Header;
}

struct egress_headers_t {
    ethernet_t          ethernet;
    ipv4_t              ipv4;
}

struct egress_metadata_t {
    bool pfIsMatch;
    pf_stats_width_t pfSessionId;
    bit<32> pfPacketLength;
}

/*************************************************************************
*********************** P A R S E R  ***********************************
*************************************************************************/

parser SwitchIngressParser(packet_in packet,
                out ingress_headers_t hdr,
                out ingress_metadata_t meta,
                out ingress_intrinsic_metadata_t ig_intr_md) {

    state start {
        /* TNA-specific Code for simple cases */
        packet.extract(ig_intr_md);
        packet.advance(PORT_METADATA_SIZE);

        meta.pfIsMatch = false;
        meta.pfSessionId = 0;
        meta.pfL4Header = {0, 0};
        meta.pfDstAddr = 0;
        meta.pfSrcAddr = 0;
        meta.pfTargetProtocol = 16w0;
        meta.origHdrLength = 0;

        transition parse_ethernet;
    }

    state parse_ethernet {
        packet.extract(hdr.ethernet);

        transition select(hdr.ethernet.etherType) {
            TYPE_IPV4: parse_ipv4;
            default: parse_post_ethernet;
        }
    }

    state parse_post_ethernet {
        meta.pfDstAddr[47:0] = hdr.ethernet.dstAddr;
        meta.pfSrcAddr[47:0] = hdr.ethernet.srcAddr;
        meta.pfTargetProtocol = hdr.ethernet.etherType;

        transition accept;
    }

    state parse_ipv4 {
        packet.extract(hdr.ipv4);

        meta.pfL4Header = packet.lookahead<l4_lookup_t>();

        meta.pfTargetProtocol[7:0] = hdr.ipv4.protocol;
        meta.pfDstAddr[31:0] = hdr.ipv4.dstAddr;
        meta.pfSrcAddr[31:0] = hdr.ipv4.srcAddr;

        transition accept;
    }

}

parser SwitchEgressParser(packet_in packet,
                out egress_headers_t hdr,
                out egress_metadata_t meta,
                out egress_intrinsic_metadata_t eg_intr_md) {

    pf_control_t pfControlHeader;

    state start {
        /* TNA-specific Code for simple cases */
        packet.extract(eg_intr_md);

        transition parse_pfControl;
    }

    state parse_pfControl {
        packet.extract(pfControlHeader);

        meta.pfIsMatch = pfControlHeader.pfIsMatch;
        meta.pfSessionId = pfControlHeader.pfSessionId;
        pfControlHeader.setInvalid();
        meta.pfPacketLength = 0;

        transition parse_ethernet;
    }

    state parse_ethernet {
        packet.extract(hdr.ethernet);
        transition select(hdr.ethernet.etherType) {
            TYPE_IPV4: parse_ipv4;
            default: accept;
        }
    }

    state parse_ipv4 {
        packet.extract(hdr.ipv4);

        transition accept;
    }
}

/*************************************************************************
**************  I N G R E S S   P R O C E S S I N G   *******************
*************************************************************************/

control SwitchIngress(inout ingress_headers_t hdr,
                  inout ingress_metadata_t meta,
                  /* Intrinsic */
                  in ingress_intrinsic_metadata_t                     ig_intr_md, 
                  in ingress_intrinsic_metadata_from_parser_t         ig_prsr_md,
                  inout ingress_intrinsic_metadata_for_deparser_t     ig_dprsr_md,
                  inout ingress_intrinsic_metadata_for_tm_t           ig_tm_md) {

    DirectCounter<bit<36>>(CounterType_t.PACKETS_AND_BYTES) pfIngressStartCounter;
    // Header byte counter after TM
    Register<bit<32>, pf_stats_width_t>(PF_TABLE_SIZE) pfIngressStartByteRegister; 
    RegisterAction<bit<32>, pf_stats_width_t, void>(pfIngressStartByteRegister) pfIngressStartByteRegisterAction = {
        void apply(inout bit<32> byteCount) {
            byteCount = byteCount + meta.origHdrLength;
        }
    };
    // Header byte counter BEFORE deparser
    Register<bit<32>, pf_stats_width_t>(PF_TABLE_SIZE) pfIngressEndByteRegister; 
    RegisterAction<bit<32>, pf_stats_width_t, void>(pfIngressEndByteRegister) pfIngressEndByteRegisterAction = {
        void apply(inout bit<32> byteCount) {
            byteCount = byteCount + sizeInBytes(hdr);
        }
    };

    action drop() {
        ig_dprsr_md.drop_ctl = 0x1; // drop packet.
    }

    action ipv4_forward(macAddr_t dstAddr, PortId_t port) {
        ig_tm_md.ucast_egress_port = port;
        hdr.ethernet.srcAddr = hdr.ethernet.dstAddr;
        hdr.ethernet.dstAddr = dstAddr;
        hdr.ipv4.ttl = hdr.ipv4.ttl - 1;
    }

    /* IPv4 routing */
    table ipv4_lpm {
        key = {
            hdr.ipv4.dstAddr: lpm;
        }
        actions = {
            ipv4_forward;
            drop;
            NoAction;
        }
        size = 16;
        default_action = drop();
    }

    /*
    sessionId is a number from 0-127 as table contains only 128 entries. Needs to be unique.
    */
    action pf_start_ingress_measure(pf_stats_width_t sessionId) {
        pfIngressStartCounter.count();
        meta.pfIsMatch = true;
        meta.pfSessionId = sessionId;
        pfIngressStartByteRegisterAction.execute(sessionId);
    }

    table pf_ig_start_selector {
        key = {
            meta.pfTargetProtocol: ternary;
            meta.pfDstAddr: ternary;
            meta.pfSrcAddr: ternary;
            meta.pfL4Header.pfLayer4Word1: ternary;
            meta.pfL4Header.pfLayer4Word2: ternary;
        }
        actions = {
            pf_start_ingress_measure;
        }
        counters = pfIngressStartCounter;
        size = 128;
    }

    action pf_end_ingress_measure() {
        pfIngressEndByteRegisterAction.execute(meta.pfSessionId);

        // Bridge metadata to egress pipeline
        hdr.pfControl.setValid();
        hdr.pfControl.pfIsMatch = meta.pfIsMatch;
        hdr.pfControl.pfSessionId = meta.pfSessionId;
    }

    apply {
        // Start Ingress measurement
        // Capture original parsed header size
        meta.origHdrLength = sizeInBytes(hdr);
        pf_ig_start_selector.apply();

        // Run IPv4 routing logic.
        ipv4_lpm.apply();

        // LAST Operation
        // End Ingress measurement.
        if (meta.pfIsMatch == true) {
            pf_end_ingress_measure();
        }
    }
}

/*************************************************************************
****************  E G R E S S   P R O C E S S I N G   *******************
*************************************************************************/

control SwitchEgress(inout egress_headers_t hdr,
                 inout egress_metadata_t meta,
                 /* Intrinsic */
                 in egress_intrinsic_metadata_t                      eg_intr_md,
                 in egress_intrinsic_metadata_from_parser_t          eg_prsr_md,
                 inout egress_intrinsic_metadata_for_deparser_t      eg_dprsr_md,
                 inout egress_intrinsic_metadata_for_output_port_t   eg_oport_md) {

    Counter<bit<36>, pf_stats_width_t>(PF_TABLE_SIZE, CounterType_t.PACKETS_AND_BYTES) pfEgressStartCounter;
    // Header byte counter BEFORE deparser
    Register<bit<32>, pf_stats_width_t>(PF_TABLE_SIZE) pfEgressEndByteRegister; 
    RegisterAction<bit<32>, pf_stats_width_t, void>(pfEgressEndByteRegister) pfEgressEndByteRegisterAction = {
        void apply(inout bit<32> byteCount) {
            byteCount = byteCount + meta.pfPacketLength;
        }
    };

    action pf_start_egress_measure() {
        // Decrement 1 byte overhead from bridge header
        pfEgressStartCounter.count(meta.pfSessionId, 1);
    }

    apply {
        // Start Egress measurement. Using count leaving TM
        if (meta.pfIsMatch == true) {
            meta.pfPacketLength = (bit<32>) eg_intr_md.pkt_length - sizeInBytes(hdr);
            pf_start_egress_measure();
        }

        // End Egress measurement. Using count from deparser
        if (meta.pfIsMatch == true) {
            meta.pfPacketLength = meta.pfPacketLength + sizeInBytes(hdr);
            pfEgressEndByteRegisterAction.execute(meta.pfSessionId);
        }
    }
}

/*************************************************************************
***********************  D E P A R S E R  *******************************
*************************************************************************/

control SwitchIngressDeparser(packet_out packet, 
                              inout ingress_headers_t hdr,
                              in ingress_metadata_t meta,
                              /* Intrinsic */
                              in ingress_intrinsic_metadata_for_deparser_t ig_dprsr_md) {

    Checksum() checksumfct;

    apply {
        //Update IPv4 checksum
        hdr.ipv4.hdrChecksum = checksumfct.update({ 
            hdr.ipv4.version,
            hdr.ipv4.ihl,
            hdr.ipv4.diffserv,
            hdr.ipv4.totalLen,
            hdr.ipv4.identification,
            hdr.ipv4.flags,
            hdr.ipv4.fragOffset,
            hdr.ipv4.ttl,
            hdr.ipv4.protocol,
            hdr.ipv4.srcAddr,
            hdr.ipv4.dstAddr 
        });

        packet.emit(hdr.pfControl);
        packet.emit(hdr.ethernet);
        packet.emit(hdr.ipv4);
    }
}

control SwitchEgressDeparser(packet_out packet,
                             inout egress_headers_t hdr,
                             in egress_metadata_t meta,
                             /* Intrinsic */
                             in egress_intrinsic_metadata_for_deparser_t eg_dprsr_md) {
    apply {
        packet.emit(hdr.ethernet);
        packet.emit(hdr.ipv4);
    }
}

/*************************************************************************
***********************  S W I T C H  *******************************
*************************************************************************/

Pipeline(
    SwitchIngressParser(), 
    SwitchIngress(), 
    SwitchIngressDeparser(), 
    SwitchEgressParser(), 
    SwitchEgress(), 
    SwitchEgressDeparser()
) pipe;
Switch(pipe) main;
