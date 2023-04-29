
// Size of lookup table in bits. 7 bits = 127 entries
#define PF_TABLE_SIZE_WIDTH 7
#define PF_TABLE_SIZE 1<<PF_TABLE_SIZE_WIDTH

typedef bit<PF_TABLE_SIZE_WIDTH> pf_stats_width_t;

header pf_control_t {
    bool pfIsMatch;
    pf_stats_width_t pfSessionId;
}

struct pf_ingress_metadata_t {
    bool pfIsMatch;
    pf_stats_width_t pfSessionId;
    bit<32> pfOrigHdrLength;
}

struct pf_egress_metadata_t {
    bool pfIsMatch;
    pf_stats_width_t pfSessionId;
    bit<32> pfPacketLength;
}
