# Cisco DNAC Input Plugin

The Cisco DNAC plugin gathers metrics about the health of the network monitored
by DNA Assurance.

## Configuration

This section contains the default TOML to configure the plugin.  You can
generate it using `telegraf --usage <plugin-name>`.

```toml
[[inputs.dnac]]
  ## Specify DNAC Base URL
  dnacbaseurl = "sandboxdnac.cisco.com"
  ## Specify Credentials
  username = "devnetuser"
  password = "Cisco123!"
  ## Debug true/false
  debug = "false"
  ## SSL Verify true/false
  sslverify = "false"
  ## Report Client Health
  clienthealth = true
  ## Report Network Health
  networkhealth = true
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
dnac_client_health, dnac_host=dnac.example.com, host=0d5e9d46cd10, site_id=global client_type_wired_score_type_poor_root_cause_dhcp_client_count=24i, client_type_wired_score_type_poor_client_count=47i, client_type_wireless_score_value=80i, client_type_wireless_score_type_poor_client_count=7i, client_type_wireless_score_type_poor_root_cause_other_client_count=3i, client_type_wireless_score_type_poor_root_cause_association_client_count=3i, client_type_wireless_score_type_fair_client_count=20i, client_type_all_score_value=91i, client_type_wired_client_count=704i, client_type_wired_score_type_poor_root_cause_aaa_client_count=10i, client_type_wired_score_type_fair_client_count=0i, client_type_wired_score_type_idle_client_count=0i, client_type_wired_score_type_nodata_client_count=0i, client_type_wired_score_type_new_client_count=0i, client_type_wireless_client_count=134i, client_type_all_client_count=838i, client_type_wired_score_type_poor_root_cause_other_client_count=13i, client_type_wireless_score_type_poor_root_cause_aaa_client_count=1i, client_type_wireless_score_type_new_client_count=0i, client_type_wireless_score_type_good_client_count=107i, client_type_wireless_score_type_idle_client_count=0i, client_type_wireless_score_type_nodata_client_count=0i, client_type_wired_score_value=93i, client_type_wired_score_type_good_client_count=657i 1635957141000000000

dnac_network_health, dnac_host=dnac.bjcc.net, host=0d5e9d46cd10, site_id=global router_good_percentage=100, ap_bad_percentage=0, ap_no_health_percentage=0, access_no_health_count=0i, core_fair_percentage=0, distribution_bad_percentage=0, router_total_count=4i, wlc_good_count=1i, core_total_count=1i, access_total_count=30i, distribution_fair_count=0i, distribution_no_health_percentage=0, router_good_count=4i, core_fair_count=0i, router_no_health_percentage=0, distribution_no_health_count=0i, core_bad_percentage=0, access_fair_count=0i, distribution_total_count=5i, distribution_good_count=5i, router_fair_percentage=0, overall_no_health_count=0, access_bad_percentage=0, router_fair_count=0i, wlc_health_score=100i, overall_health_score=87i, access_fair_percentage=0, router_health_score=100i, wlc_total_count=1i, wlc_bad_count=0i, wlc_no_health_percentage=0, ap_health_score=85i, ap_good_count=228i, overall_fair_count=40i, ap_fair_percentage=14.869888, ap_good_percentage=84.75836, router_bad_count=0i, wlc_no_health_count=0i, access_bad_count=0i, distribution_bad_count=0i, distribution_health_score=100i, wlc_fair_percentage=0, ap_total_count=269i, overall_good_count=269i, ap_no_health_count=0i, router_no_health_count=0i, core_health_score=100i, core_no_health_count=0i, distribution_fair_percentage=0, wlc_fair_count=0i, wlc_good_percentage=100, ap_fair_count=40i, overall_bad_count=0, core_good_count=1i, access_good_percentage=100, distribution_good_percentage=100, ap_bad_count=0i, core_bad_count=0i, access_health_score=100i, access_good_count=30i, access_no_health_percentage=0, wlc_bad_percentage=0, core_good_percentage=100, core_no_health_percentage=0, router_bad_percentage=0, overall_total_count=310i 1635957141000000000
```
