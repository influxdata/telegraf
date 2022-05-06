# Cisco DNAC Input Plugin

The Cisco DNAC plugin gathers metrics about the health of the network monitored
by DNA Assurance.

## Configuration

This section contains the default TOML to configure the plugin.  You can
generate it using `telegraf --usage <plugin-name>`.

```toml
[[inputs.cisco_dna_center]]
  ## Specify DNAC Base URL
  dnacbaseurl = "sandboxdnac.cisco.com"
  ## Specify Credentials
  username = "devnetuser"
  password = "Cisco123!"
  ## SSL Verify true/false
  sslverify = "false"
  ## Health types to report
  report = ["client","network"]
```

## Metrics

Client Health API -
[api doc](https://developer.cisco.com/docs/dna-center/#!get-overall-client-health)

- dnac_client_health
  - tags:
    - dnac_host
    - site_id
  - fields:
    - client_type_\<type\>_client_count (type, unit)
    - client_type_\<type\>_score_value (type, unit)
    - client_type_\<type\>_score_type\_\<type\>_client_count (type, unit)
    - client_type_\<type\>_score_type\_\<type>_root_cause\_\<cause\>_client_count (type, unit)

Network Health API -
[api doc](https://developer.cisco.com/docs/dna-center/#!get-overall-network-health)

- dnac_network_health
  - tags:
    - dnac_host
    - site_id
  - fields:
    - overall_health_score (type, uint)
    - overall_total_count (type, uint)
    - overall_no_health_count (type, uint)
    - overall_good_count (type, uint)
    - overall_fair_count (type, uint)
    - overall_bad_count (type, uint)
    - \<device_type\>_health_score (type, uint)
    - \<device_type\>_total_count (type, uint)
    - \<device_type\>_bad_count (type, uint)
    - \<device_type\>_bad_percentage (type, float)
    - \<device_type\>_fair_count (type, uint)
    - \<device_type\>_fair_percentage (type, float)
    - \<device_type\>_good_count (type, uint)
    - \<device_type\>_good_percentage (type, float)
    - \<device_type\>_no_health_count (type, uint)
    - \<device_type\>_no_health_percentage (type, float)

### Example Output

This section shows example output in Line Protocol format.  You can often use
`telegraf --input-filter dnac --test` or use the `file` output to get
this information.

```sh
dnac_client_health,host=sandboxdnac.cisco.com,site_id=global client_type_wired_score_type_idle_client_count=0i,client_type_wired_score_type_new_client_count=0i,client_type_wireless_client_count=0i,client_type_wireless_score_type_fair_client_count=0i,client_type_wireless_score_type_good_client_count=0i,client_type_wireless_score_value=-1i,client_type_wireless_score_type_idle_client_count=0i,client_type_wireless_score_type_nodata_client_count=0i,client_type_wireless_score_type_new_client_count=0i,client_type_all_score_value=-1i,client_type_wired_client_count=0i,client_type_wired_score_type_good_client_count=0i,client_type_wired_score_type_nodata_client_count=0i,client_type_all_client_count=0i,client_type_wired_score_value=-1i,client_type_wired_score_type_poor_client_count=0i,client_type_wired_score_type_fair_client_count=0i,client_type_wireless_score_type_poor_client_count=0i 1644440389000000000

dnac_network_health,host=sandboxdnac.cisco.com,site_id=global access_health_score=100i,access_bad_percentage=0,access_fair_count=0,distribution_good_percentage=100,wlc_bad_count=0,wlc_bad_percentage=0,overall_fair_count=0i,distribution_health_score=100i,distribution_bad_count=0,distribution_fair_count=0,wlc_fair_percentage=0,access_bad_count=0,distribution_total_count=1i,distribution_bad_percentage=0,distribution_fair_percentage=0,overall_bad_count=0,overall_health_score=75i,access_fair_percentage=0,wlc_fair_count=0,wlc_good_count=0i,access_total_count=2i,access_good_count=2i,access_good_percentage=100,distribution_good_count=1i,wlc_total_count=1i,overall_total_count=4i,overall_good_count=3i,wlc_health_score=0i,wlc_good_percentage=0 1644440376000000000
```
