/* -*- P4_16 -*- */
#include <core.p4>
/* TOFINO Native architecture */
#include <t2na.p4>

#include "include/pifina_headers.p4"

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

header tcp_t {
    bit<16> src_port;
    bit<16> dst_port;
}

header udp_t {
    bit<16> src_port;
    bit<16> dst_port;
}

struct ingress_headers_t {
    pf_control_t        pfControl;
    ethernet_t          ethernet;
    ipv4_t              ipv4;
}

struct ingress_metadata_t {
    pf_ingress_metadata_t pf_meta;
}

struct egress_headers_t {
    ethernet_t          ethernet;
    ipv4_t              ipv4;
}

struct egress_metadata_t {
    pf_egress_metadata_t pf_meta;
}

#include "include/pifina_probes.p4"

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

        meta.pf_meta.pfIsMatch = false;
        meta.pf_meta.pfSessionId = 0;
        meta.pf_meta.pfOrigHdrLength = 0;

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

        meta.pf_meta.pfIsMatch = pfControlHeader.pfIsMatch;
        meta.pf_meta.pfSessionId = pfControlHeader.pfSessionId;
        pfControlHeader.setInvalid();
        meta.pf_meta.pfPacketLength = 0;

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

    PfIngressStartProbe() pfIngressStartProbe;
    PfIngressEndProbe() pfIngressEndProbe;

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

    apply {
        // Start Ingress measurement
        pfIngressStartProbe.apply(hdr, meta);

        // Run IPv4 routing logic.
        ipv4_lpm.apply();

        // LAST Operation
        pfIngressEndProbe.apply(hdr, meta);
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

    PfEgressStartProbe() pfEgressStartProbe;
    PfEgressEndProbe() pfEgressEndProbe;

    apply {
        // Start Egress measurement. Using count leaving TM
        pfEgressStartProbe.apply(hdr, meta, eg_intr_md);

        // End Egress measurement. Using count from deparser
        pfEgressEndProbe.apply(hdr, meta);
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
