# Performance Introspector for in-network applications (PIFINA)
Authors:
* Thushjandan Ponnudurai

## Introduction
This is a performance framework to introspect in-network applications, which are written in P4 programming language. It has been developed and tested for the Intel Tofino architecture 2, but it is backwards compatible to Tofino 1.

## Snippets
```
bfrt.pifina.pipe.SwitchIngress.pf_ig_start_selector.add_with_pf_start_ingress_measure(pfTargetProtocol=0xFA,pfTargetProtocol_mask=0xFF,pfDstAddr=0x0A000202,pfDstAddr_mask=0xFFFFFFFF,pfSrcAddr=0x0A000101,pfSrcAddr_mask=0xFFFFFFFF, pfLayer4Word1=0x0,pfLayer4Word1_mask=0x0, pfLayer4Word2=0x0, pfLayer4Word2_mask=0x0, sessionId=0x4)
```
