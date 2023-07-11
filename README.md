# Performance Introspector for in-network applications (PIFINA)
Authors:
* Thushjandan Ponnudurai

## Introduction
This is a performance framework to introspect in-network applications, which are written in P4 programming language. It has been developed and tested for the Intel Tofino architecture 2, but it is backwards compatible to Tofino 1.

## Snippets
```
# Tofino
bfrt.pifina.pipe.SwitchIngress.pfIngressStartProbe.PF_INGRESS_MATCH_CNT.add_with_pf_start_ingress_measure(protocol=0xFA,dstAddr=0x0a00030e,dstAddr_mask=0xFFFFFFFF,srcAddr=0x0a00030d,srcAddr_mask=0xFFFFFFFF, sessionId=0x7)
# Emulator
bfrt.pifina.pipe.SwitchIngress.pfIngressStartProbe.PF_INGRESS_MATCH_CNT.add_with_pf_start_ingress_measure(protocol=0xFA,dstAddr=0x0A000202,dstAddr_mask=0xFFFFFFFF,srcAddr=0x0A000102,srcAddr_mask=0xFFFFFFFF, sessionId=0x6)
bfrt.pifina.pipe.SwitchIngress.pfIngressStartProbe.PF_INGRESS_MATCH_CNT.add_with_pf_start_ingress_measure(protocol=0xFA,dstAddr=0x0A000202,dstAddr_mask=0xFFFFFFFF,srcAddr=0x0A000101,srcAddr_mask=0xFFFFFFFF, sessionId=0x4)
```
