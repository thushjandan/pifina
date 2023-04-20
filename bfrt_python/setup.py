from collections import namedtuple

P4_PROG = "pf_poc" # replace with the current P4 program

print("Starting setup for {prog}".format(prog=P4_PROG))

exec("p4 = bfrt.{prog}.pipe".format(prog=P4_PROG))  # access to the program-specific table config
port = bfrt.port                                    # port management
pre = bfrt.pre                                      # packet replication engine

port_map = [ # port_name | mac | speed
    # Pipeline 1
    ("1/0", "00:00:00:00:00:01", "BF_SPEED_50G"),
    ("2/0", "00:00:00:00:00:02", "BF_SPEED_50G"),
    ("3/0", "00:00:00:00:00:03", "BF_SPEED_50G"),
    ("4/0", "00:00:00:00:00:04", "BF_SPEED_50G"),
]

ipv4_routes = [
    {"dest": "10.0.1.1", "length": 32, "dest_mac": "00:00:00:00:00:01", "port": "1/0"},
    {"dest": "10.0.2.2", "length": 32, "dest_mac": "00:00:00:00:00:02", "port": "2/0"}
]

SwitchPort = namedtuple("SwitchPort", [ "port_name", "mac", "speed" ])
port_map = list(map(lambda tuple: SwitchPort._make(tuple), port_map))

### ========================================================================================================================================== ###

print("Clearing table entries")

def clear_all(verbose=True, batching=True):
    global p4
    global bfrt
    
    def _clear(table, verbose=False, batching=False):
        if verbose:
            print("Clearing table {:<40} ... ".format(table['full_name']), end='', flush=True)
        try:    
            entries = table['node'].get(regex=True, print_ents=False)
            try:
                if batching: bfrt.batch_begin()

                for entry in entries:
                    entry.remove()
            except Exception as e:
                print("Problem clearing table {}: {}".format(table['name'], e.sts))
            finally:
                if batching: bfrt.batch_end()
        except Exception as e:
            if e.sts == 6 and verbose: print('(Empty) ', end='')
        finally:
            if verbose: print('Done')

        try: table['node'].reset_default() # reset the default action if there is any
        except: pass

    # The order is important. We do want to clear from the top, table entries use selector groups and selector groups use action profile members    
    for table in p4.info(return_info=True, print_info=False): # Clear Match Tables
        if "db_join" in table['full_name'] or "db_drop_reply" in table['full_name']:
            continue
        if table['type'] in ['MATCH_DIRECT', 'MATCH_INDIRECT_SELECTOR']:
            _clear(table, verbose=verbose, batching=batching)

    for table in p4.info(return_info=True, print_info=False): # Clear Selectors
        if table['type'] in ['SELECTOR']:
            _clear(table, verbose=verbose, batching=batching)
            
    for table in p4.info(return_info=True, print_info=False): # Clear Action Profiles
        if table['type'] in ['ACTION_PROFILE']:
            _clear(table, verbose=verbose, batching=batching)

clear_all()

### ========================================================================================================================================== ###

### Returns device_port given a port_name under the form "front-port/lane", e.g. "1/0" returns "136" ###
def get_device_port(port_name):
    return port.port_str_info.get(PORT_NAME=port_name, return_ents=True, print_ents=False).data[b'$DEV_PORT'] # port_str_info.get() returns a TableEntry object with .key and .data members

### Returns how many lanes a connection needs given a speed ###
SPEED_TO_LANES = { 'BF_SPEED_50G': 1, 'BF_SPEED_100G': 2, 'BF_SPEED_200G': 4, 'BF_SPEED_400G': 8 }
def get_lanes(speed): 
    return SPEED_TO_LANES[speed]

### Clears a single configuration table ###
def clear_table(table):
    try:
        for entry in table.get(regex=True, print_ents=False): entry.remove()
    except: pass

### ========================================================================================================================================== ###

print("Populating table entries")

# Forward Table
for ip_entry in ipv4_routes:
    p4.SwitchIngress.ipv4_lpm.add_with_ipv4_forward(
        hdr_ipv4_dstAddr=ip_entry["dest"],
        hdr_ipv4_dstAddr_p_length=ip_entry["length"],
        dstAddr=ip_entry["dest_mac"],
        port=get_device_port(ip_entry["port"])
    )

### ========================================================================================================================================== ###

print("Configuring Tofino 2")

port.port.clear()       # clears all current port configurations
clear_table(pre.mgid)   # clears all current multicast groups
clear_table(pre.node)   # clears all current multicast nodes

print("Configuring ports") # port.port.string_choices() # prints all possible options for port.port.add()

for switch_port in port_map: # Adding and enabling ports
    port.port.add(
        DEV_PORT=get_device_port(switch_port.port_name),
        SPEED=switch_port.speed, N_LANES=get_lanes(switch_port.speed),
        FEC='BF_FEC_TYP_REED_SOLOMON',
        AUTO_NEGOTIATION='PM_AN_FORCE_DISABLE',
        LOOPBACK_MODE="BF_LPBK_NONE",
        PORT_ENABLE=True,
    ) 

### ========================================================================================================================================== ###

bfrt.complete_operations()
