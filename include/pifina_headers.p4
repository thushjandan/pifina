
/**
* PIFINA P4 headers
*
* Copyright 2023 Thushjandan Ponnudurai

* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*/


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
