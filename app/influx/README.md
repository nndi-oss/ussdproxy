InfluxDB Forwarder
==================

This application is a Proof of concept for a uni-directional
influxdb data receiver via UDCP.
Basically, we receive data from clients and send that data
to an influxdb server, nothing more, nothing less.


## Example Usage

```
*386*8327*UDCP_INIT_PROTOCOL    v:1.0.0# 
*386*8327*UDCP_INIT_APP         app:influx,sep:pipe,client:12345# 
*386*8327*UDCP_DATA_MTS         temp:23.7|tag_meter_no:abcdef|heat:100|tag_heat_si:DEG# 
*386*8327*UDCP_DATA_MTS         temp:23.7|tag_meter_no:abcdef|heat:100|tag_heat_si:DEG# 
*386*8327*UDCP_DATA_MTS         temp:23.7|tag_meter_no:abcdef|heat:100|tag_heat_si:DEG# 
*386*8327*UDCP_DATA_MTS         temp:23.7|tag_meter_no:abcdef|heat:100|tag_heat_si:DEG# 
*386*8327*UDCP_DATA             temp:23.7|tag_meter_no:abcdef|heat:100|tag_heat_si:DEG# 
*386*8327*UDCP_RELEASE_DIALOGUE app:influx client:12345# 
```