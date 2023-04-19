/* -*- P4_16 -*- */
#include <core.p4>
/* TOFINO Native architecture */
#include <t2na.p4>

const bit<16> TYPE_IPV4 = 0x800;

/*************************************************************************
*********************** H E A D E R S  ***********************************
*************************************************************************/

typedef bit<48> macAddr_t;
typedef bit<32> ip4Addr_t;

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

header tcp_h {
    bit<16> src_port;
    bit<16> dst_port;
}

header udp_h {
    bit<16> src_port;
    bit<16> dst_port;
}

struct l4_lookup_t {
    bit<16> pfLayer4Word1; // Layer 4 destination port
    bit<16> pfLayer4Word2; // Layer 4 source port
}

struct metadata {
    bit<1> pfIsMatch;
    bit<16> pfTargetProtocol;
    bit<128> pfDstAddr;
    bit<128> pfSrcAddr;
    l4_lookup_t pfL4Header;
}

struct headers {
    ethernet_t          ethernet;
    ipv4_t              ipv4;
}

/*************************************************************************
*********************** P A R S E R  ***********************************
*************************************************************************/

parser SwitchIngressParser(packet_in packet,
                out headers hdr,
                out metadata meta,
                out ingress_intrinsic_metadata_t ig_intr_md) {

    state start {
        /* TNA-specific Code for simple cases */
        packet.extract(ig_intr_md);
        packet.advance(PORT_METADATA_SIZE);

        meta.pfL4Header = {0, 0};
        meta.pfDstAddr = 0;
        meta.pfSrcAddr = 0;

        transition parse_ethernet;
    }

    state parse_ethernet {
        packet.extract(hdr.ethernet);

        meta.pfDstAddr = (bit<128>)hdr.ethernet.dstAddr;
        meta.pfSrcAddr = (bit<128>)hdr.ethernet.srcAddr;
        meta.pfTargetProtocol = hdr.ethernet.etherType;

        transition select(hdr.ethernet.etherType) {
            TYPE_IPV4: parse_ipv4;
            default: accept;
        }
    }

    state parse_ipv4 {
        packet.extract(hdr.ipv4);

        meta.pfDstAddr = (bit<128>)hdr.ipv4.dstAddr;
        meta.pfSrcAddr = (bit<128>)hdr.ipv4.srcAddr;
        meta.pfTargetProtocol = 0;
        meta.pfTargetProtocol = (bit<16>)hdr.ipv4.protocol;
        meta.pfL4Header = packet.lookahead<l4_lookup_t>();

        transition accept;
    }

}

parser SwitchEgressParser(packet_in packet,
                out headers hdr,
                out metadata meta,
                out egress_intrinsic_metadata_t eg_intr_md) {

    state start {
        /* TNA-specific Code for simple cases */
        packet.extract(eg_intr_md);

        meta.pfL4Header = {0, 0};
        meta.pfDstAddr = 0;
        meta.pfSrcAddr = 0;
        meta.pfTargetProtocol = 0;

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
        meta.pfL4Header = packet.lookahead<l4_lookup_t>();
        transition accept;
    }
}

/*************************************************************************
**************  I N G R E S S   P R O C E S S I N G   *******************
*************************************************************************/

control SwitchIngress(inout headers hdr,
                  inout metadata meta,
                  /* Intrinsic */
                  in ingress_intrinsic_metadata_t                     ig_intr_md, 
                  in ingress_intrinsic_metadata_from_parser_t         ig_prsr_md,
                  inout ingress_intrinsic_metadata_for_deparser_t     ig_dprsr_md,
                  inout ingress_intrinsic_metadata_for_tm_t           ig_tm_md) {
    
    DirectCounter<bit<36>>(CounterType_t.PACKETS_AND_BYTES) pfIngressStartCounter;
    DirectCounter<bit<36>>(CounterType_t.PACKETS_AND_BYTES) pfIngressEndCounter;    

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
    action pf_start_measure(bit<7> sessionId) {
        
        pfIngressStartCounter.count();

    }

    table pf_ig_start_selector {
        key = {
            meta.pfTargetProtocol: exact;
            meta.pfDstAddr: ternary;
            meta.pfSrcAddr: ternary;
            meta.pfL4Header.pfLayer4Word1: ternary;
            meta.pfL4Header.pfLayer4Word2: ternary;
        }
        actions = {
            pf_start_measure;
        }
        counters = pfIngressStartCounter;
        size = 128;
    }

    action pf_end_measure() {
        pfIngressEndCounter.count();

    }

    table pf_ig_end_selector {
        key = {
            meta.pfTargetProtocol: exact;
            meta.pfDstAddr: ternary;
            meta.pfSrcAddr: ternary;
            meta.pfL4Header.pfLayer4Word1: ternary;
            meta.pfL4Header.pfLayer4Word2: ternary;
        }

        actions = {
            pf_end_measure;
        }

        counters = pfIngressEndCounter;
        size = 128;
    }

    apply {
        // Start Ingress measurement
        pf_ig_start_selector.apply();

        // Run IPv4 routing logic.
        ipv4_lpm.apply();

        // End Ingress measurement
        pf_ig_end_selector.apply();
    }
}

/*************************************************************************
****************  E G R E S S   P R O C E S S I N G   *******************
*************************************************************************/

control SwitchEgress(inout headers hdr,
                 inout metadata meta,
                 /* Intrinsic */
                 in egress_intrinsic_metadata_t                      eg_intr_md,
                 in egress_intrinsic_metadata_from_parser_t          eg_prsr_md,
                 inout egress_intrinsic_metadata_for_deparser_t      eg_dprsr_md,
                 inout egress_intrinsic_metadata_for_output_port_t   eg_oport_md) {

    apply {

    }
}

/*************************************************************************
***********************  D E P A R S E R  *******************************
*************************************************************************/

control SwitchIngressDeparser(packet_out packet, 
                              inout headers hdr,
                              in metadata meta,
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

        packet.emit(hdr.ethernet);
        packet.emit(hdr.ipv4);
    }
}

control SwitchEgressDeparser(packet_out packet,
                             inout headers hdr,
                             in metadata meta,
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
