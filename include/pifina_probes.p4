control PfIngressStartProbe(in ingress_headers_t hdr, inout ingress_metadata_t meta) {
    DirectCounter<bit<36>>(CounterType_t.PACKETS_AND_BYTES) pfIngressStartCounter;
    // Header byte counter after TM
    @name("PF_INGRESS_START_HDR_SIZE")
    Register<bit<64>, pf_stats_width_t>(PF_TABLE_SIZE) pfIngressStartByteRegister; 
    RegisterAction<bit<64>, pf_stats_width_t, void>(pfIngressStartByteRegister) pfIngressStartByteRegisterAction = {
        void apply(inout bit<64> byteCount) {
            byteCount = byteCount + (bit<64>) meta.pf_meta.pfOrigHdrLength;
        }
    };

    /*
    sessionId is a number from 0-127 as table contains only 128 entries. Needs to be unique.
    */
    action pf_start_ingress_measure(pf_stats_width_t sessionId) {
        pfIngressStartCounter.count();
        meta.pf_meta.pfIsMatch = true;
        meta.pf_meta.pfSessionId = sessionId;
        pfIngressStartByteRegisterAction.execute(sessionId);
    }

    @name("PF_INGRESS_MATCH_CNT")
    table pf_ig_start_selector {
        key = {
            hdr.ipv4.protocol: exact;
            hdr.ipv4.dstAddr: ternary;
            hdr.ipv4.srcAddr: ternary;
        }
        actions = {
            pf_start_ingress_measure;
        }
        counters = pfIngressStartCounter;
        size = 128;
    }

    apply {
        // Start Ingress measurement
        // Capture original parsed header size
        meta.pf_meta.pfOrigHdrLength = sizeInBytes(hdr);
        pf_ig_start_selector.apply();
    }
}

control PfIngressEndProbe(inout ingress_headers_t hdr, in ingress_metadata_t meta) {
    // Header byte counter BEFORE deparser
    @name("PF_INGRESS_END_HDR_SIZE")
    Register<bit<64>, pf_stats_width_t>(PF_TABLE_SIZE) pfIngressEndByteRegister; 
    RegisterAction<bit<64>, pf_stats_width_t, void>(pfIngressEndByteRegister) pfIngressEndByteRegisterAction = {
        void apply(inout bit<64> byteCount) {
            byteCount = byteCount + (bit<64>) sizeInBytes(hdr);
        }
    };

    action pf_end_ingress_measure() {
        pfIngressEndByteRegisterAction.execute(meta.pf_meta.pfSessionId);

        // Bridge metadata to egress pipeline
        hdr.pfControl.setValid();
        hdr.pfControl.pfIsMatch = meta.pf_meta.pfIsMatch;
        hdr.pfControl.pfSessionId = meta.pf_meta.pfSessionId;
    }

    apply {
        // LAST Operation
        // End Ingress measurement.
        if (meta.pf_meta.pfIsMatch == true) {
            pf_end_ingress_measure();
        }
    }
}

control PfEgressStartProbe(in egress_headers_t hdr, inout egress_metadata_t meta, in egress_intrinsic_metadata_t eg_intr_md) {
    @name("PF_EGRESS_START_CNT")
    Counter<bit<36>, pf_stats_width_t>(PF_TABLE_SIZE, CounterType_t.PACKETS_AND_BYTES) pfEgressStartCounter;

    action pf_start_egress_measure() {
        // Decrement 1 byte overhead from bridge header
        pfEgressStartCounter.count(meta.pf_meta.pfSessionId, 1);
    }
    
    apply {
        if (meta.pf_meta.pfIsMatch == true) {
            meta.pf_meta.pfPacketLength = (bit<32>) eg_intr_md.pkt_length - sizeInBytes(hdr);
            pf_start_egress_measure();
        }
    }
}

control PfEgressEndProbe(in egress_headers_t hdr, inout egress_metadata_t meta) {
    // Header byte counter BEFORE deparser
    @name("PF_EGRESS_END_CNT")
    Register<bit<32>, pf_stats_width_t>(PF_TABLE_SIZE) pfEgressEndByteRegister; 
    RegisterAction<bit<32>, pf_stats_width_t, void>(pfEgressEndByteRegister) pfEgressEndByteRegisterAction = {
        void apply(inout bit<32> byteCount) {
            byteCount = byteCount + meta.pf_meta.pfPacketLength;
        }
    };
    
    apply {
        if (meta.pf_meta.pfIsMatch == true) {
            meta.pf_meta.pfPacketLength = meta.pf_meta.pfPacketLength + sizeInBytes(hdr);
            pfEgressEndByteRegisterAction.execute(meta.pf_meta.pfSessionId);
        }
    }
}