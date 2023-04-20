#!/usr/bin/env python3
import sys

from scapy.all import (
    IntField,
    IP,
    UDP,
    bind_layers,
    Packet,
    BitField,
    sniff,
)

class DBEntry(Packet):
    fields_desc = [ 
        IntField("entryId", 0),
        IntField("secondAttr", 0),
        IntField("thirdAttr", 0),
    ]

class DBReplyEntry(Packet):
    fields_desc = [ 
        IntField("entryId", 0),
        IntField("secondAttr", 0),
        IntField("thirdAttr", 0),
        IntField("forthAttr", 0),
        IntField("fifthAttr", 0),
    ]

class DBRelation(Packet):
    name = "MYP4DB_Relation"
    fields_desc = [ 
        BitField("relationId", 0, 8),
        BitField("joinedRelationId", 0, 8),
    ]

# IP proto 250 indicates MYP4DB_Relation
bind_layers(IP, DBRelation, proto=0xFA)
# If isReply is set, then it is a reply packet.
bind_layers(DBRelation, DBReplyEntry)
# If bottom of stack has reached, then UDP header will follow
bind_layers(DBEntry, UDP)
# If bottom of stack has reached, then UDP header will follow
bind_layers(DBReplyEntry, UDP)

def handle_pkt(pkt):
    print("got a packet")
    pkt.show2()
    sys.stdout.flush()


def main():
    iface = 'veth17'
    print("sniffing on %s" % iface)
    sys.stdout.flush()
    # Listen on MYP4DB_Relation and UDP packets
    sniff(filter="proto (250 or 17)", iface = iface,
          prn = lambda x: handle_pkt(x))

if __name__ == '__main__':
    main()