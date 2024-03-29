/* -*- P4_16 -*- */
/* TOFINO Native architecture */
#include <core.p4>
#include <tna.p4>

// PIFINA: Step 1: Include header files
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

// PIFINA: Step 2: Include PIFINA functions
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

        // PIFINA: Step 3. Call pifina_ingress_parser with pf_meta as parameter 
        pifina_ig_parser_init(meta.pf_meta);

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

    state start {
        /* TNA-specific Code for simple cases */
        packet.extract(eg_intr_md);

        // PIFINA: Step 5: Execute Pifina Egress Parser
        pifina_eg_parser_init(packet, meta.pf_meta);

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

    // PIFINA: Step 6: Initialize Pifina control blocks
    PfIngressStartProbe() pfIngressStartProbe;
    PfIngressEndProbe() pfIngressEndProbe;
    {{ range .ExtraProbeList }}
    {{- if eq .Type "INGRESS" }}
    PfIngressExtraProbe{{ .Name }}() pfIngressExtraProbe{{ .Name }};
    {{- end }}
    {{end}}

    apply {
        // PIFINA: Step 7: Start Ingress measurement
        pfIngressStartProbe.apply(hdr, meta.pf_meta);

        // YOUR CODE COMES HERE
        {{ range .ExtraProbeList }}
        {{- if eq .Type "INGRESS" }}
        pfIngressExtraProbe{{ .Name }}.apply(hdr, meta.pf_meta);
        {{- end }}
        {{end}}

        // PIFINA: Step 8: last Operation
        pfIngressEndProbe.apply(hdr, meta.pf_meta, ig_intr_md);
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

    // PIFINA: Step 9: Initialize PIFINA probes
    PfEgressStartProbe() pfEgressStartProbe;
    PfEgressEndProbe() pfEgressEndProbe;
    {{ range .ExtraProbeList }}
    {{- if eq .Type "EGRESS" }}
    PfEgressExtraProbe{{ .Name }}() pfEgressExtraProbe{{ .Name }};
    {{- end }}
    {{end}}

    apply {
        // PIFINA: Step 9: Start Egress measurement. Using count leaving TM
        pfEgressStartProbe.apply(hdr, meta.pf_meta, eg_intr_md);

        // YOUR CODE COMES HERE

        // PIFINA Extra Probes to use
        {{ range .ExtraProbeList }}
        {{- if eq .Type "EGRESS" }}
        pfEgressExtraProbe{{ .Name }}.apply(hdr, meta.pf_meta);
        {{- end }}
        {{end}}

        // PIFINA: Step 10: End Egress measurement. Using count from deparser
        pfEgressEndProbe.apply(hdr, meta.pf_meta);
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

        // PIFINA: Step 11: Emit Pifina bridge header
        packet.emit(meta.pf_meta.pfControl);
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
