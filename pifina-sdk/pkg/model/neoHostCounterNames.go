// Copyright (c) 2023 Thushjandan Ponnudurai
// 
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package model

const (
	NEOHOST_TPT_L0_MTT_HIT                   = "Level 0 MTT Cache Hit"
	NEOHOST_TPT_L0_MTT_MISS                  = "Level 0 MTT Cache Miss"
	NEOHOST_TPT_L1_MTT_HIT                   = "Level 1 MTT Cache Hit"
	NEOHOST_TPT_L1_MTT_MISS                  = "Level 1 MTT Cache Miss"
	NEOHOST_TPT_L0_MPT_HIT                   = "Level 0 MPT Cache Hit"
	NEOHOST_TPT_L0_MPT_MISS                  = "Level 0 MPT Cache Miss"
	NEOHOST_TPT_L1_MPT_HIT                   = "Level 1 MPT Cache Hit"
	NEOHOST_TPT_L1_MPT_MISS                  = "Level 1 MPT Cache Miss"
	NEOHOST_PXT_PXD_READY_BP                 = "PCIe Internal Back Pressure"
	NEOHOST_PXT_PERF_RD_ICMC_PUSH_LINK0      = "ICM Cache Miss"
	NEOHOST_RXB_BUFFER_FULL_PERF_COUNT_PORT0 = "RX Packet Buffer Full Port 0"
	NEOHOST_RXB_BUFFER_FULL_PERF_COUNT_PORT1 = "RX Packet Buffer Full Port 1"
	NEOHOST_RXW_PERF_WB_HIT                  = "Receive WQE Cache Hit"
	NEOHOST_RXW_PERF_WB_MISS                 = "Receive WQE Cache Miss"
	NEOHOST_TX_BW                            = "TX BandWidth"
	NEOHOST_RX_BW                            = "RX BandWidth"
	NEOHOST_TX_PACKET_RATE                   = "TX Packet Rate"
	NEOHOST_RX_PACKET_RATE                   = "RX Packet Rate"
	NEOHOST_PCI_INBOUND_BW                   = "PCIe Inbound BW Utilization"
	NEOHOST_PCI_OUTBOUND_BW                  = "PCIe Outbound BW Utilization"
)

var NEOHOST_COUNTERS = []string{
	NEOHOST_TPT_L0_MTT_HIT,
	NEOHOST_TPT_L0_MTT_MISS,
	NEOHOST_TPT_L1_MTT_HIT,
	NEOHOST_TPT_L1_MTT_MISS,
	NEOHOST_TPT_L0_MPT_HIT,
	NEOHOST_TPT_L0_MPT_MISS,
	NEOHOST_TPT_L1_MPT_HIT,
	NEOHOST_TPT_L1_MPT_MISS,
	NEOHOST_PXT_PXD_READY_BP,
	NEOHOST_PXT_PERF_RD_ICMC_PUSH_LINK0,
	NEOHOST_RXB_BUFFER_FULL_PERF_COUNT_PORT0,
	NEOHOST_RXB_BUFFER_FULL_PERF_COUNT_PORT1,
	NEOHOST_RXW_PERF_WB_HIT,
	NEOHOST_RXW_PERF_WB_MISS,
	NEOHOST_TX_BW,
	NEOHOST_RX_BW,
	NEOHOST_TX_PACKET_RATE,
	NEOHOST_RX_PACKET_RATE,
	NEOHOST_PCI_INBOUND_BW,
	NEOHOST_PCI_OUTBOUND_BW,
}
