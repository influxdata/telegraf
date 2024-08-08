# ah_wireless Input Plugin

The `ah_wireless` plugin gather metrics on the wireless statistics from HiveOS.

## Configuration

```toml
# Read metrics about wireless stats
[[inputs.ah_wireless]]
  # Sample interval
  interval = "5s"
  # Interface names from where stats to be collected
  ifname = ["wifi0","wifi1"]
```

## Metrics

There are two Metrics:
RfStats
and
ClientStats

- RfStats
  - device:
    - serialnumber (string)
  - items:
    - channelUtilization:
    - interferenceUtilization:

  txUtilization:

  rxUtilization:

  rxInbssUtilization:

  rxObssUtilization:

  wifinterferenceUtilization:

  noise:

  crcErrorRate:

  txPackets:
    description: Counter - Number of transmitted packets.
    type: integer
    format: int64
  txErrors:
    description: Counter - Number of transmitted packets with errors.
    type: integer
    format: int64
  txDropped:
    description: Counter - Number of dropped transmit packets.
    type: integer
    format: int64
  txHwDropped:
    description: Counter - Number of dropped transmit packets by HW.
    type: integer
    format: int64
  txSwDropped:
    description: Counter - Number of dropped transmit packets by SW.
    type: integer
    format: int64
  txBytes:
    description: Counter - Number of transmitted bytes.
    type: integer
    format: int64
  txRetryCount:
    description: Counter - Number of transmit retries.
    type: integer
    format: int64
  txRate:
    $ref: "../../../common/components/schemas/FloatGauge.yaml"
  txUnicastPackets:
    description: Counter - Number of transmitted Unicast packets.
    type: integer
    format: int64
  txMulticastPackets:
    description: Counter - Number of transmitted Multicast packets.
    type: integer
    format: int64
  txMulticastBytes:
    description: Counter - Number of transmitted Multicast bytes.
    type: integer
    format: int64
  txBcastPackets:
    description: Counter - Number of transmitted Broadcast packets.
    type: integer
    format: int64
  txBcastBytes:
    description: Counter - Number of transmitted Broadcast bytes.
    type: integer
    format: int64
  rxPackets:
    description: Counter - Number of received packets.
    type: integer
    format: int64
  rxErrors:
    description: Counter - Number of received packets with errors.
    type: integer
    format: int64
  rxDropped:
    description: Counter - Number of dropped receive packets.
    type: integer
    format: int64
  rxBytes:
    description: Counter - Number of received bytes.
    type: integer
    format: int64
  rxRetryCount:
    description: Counter - Number of receive retries.
    type: integer
    format: int64
  rxRate:

  rxUnicastPackets:
    description: Counter - Number of received Unicast packets.
    type: integer
    format: int64
  rxMulticastPackets:
    description: Counter - Number of received Multicast packets.
    type: integer
    format: int64
  rxMulticastBytes:
    description: Counter - Number of received Multicast bytes.
    type: integer
    format: int64
  rxBcastPackets:
    description: Counter - Number of received Broadcast packets.
    type: integer
    format: int64
  rxBcastBytes:
    description: Counter - Number of received Broadcast bytes.
    type: integer
    format: int64
  bsSpCnt:
    description: Band steering suppress count.
    type: integer
    format: int32
  lbSpCnt:
    description: Load balance suppress count.
    type: integer
    format: int32
  snrSpCnt:
    description: Weak snr suppress count.
    type: integer
    format: int32
  snAnswerCnt:
    description: Safety net answer (safety net check fail) count.
    type: integer
    format: int32
  rxPrbSpCnt:
    description: Probe request suppressed count .
    type: integer
    format: int32
  rxAuthCnt:
    description: Auth request suppressed count.
    type: integer
    format: int32
  txBitrateSuc:
    description: Total TX bit rate success distribution percentage.
    type: integer
    format: int8
  rxBitrateSuc:
    description: Total RX bit rate success distribution percentage.
    type: integer
    format: int8
  rxRateStats:
    description: rx distribution bit rate
    type: array

  txRateStats:
    description: tx distribution bit rate
    type: array

  clientCount:

ClientStats:
description: |-
  assocTime: Wireless association time in milliseconds.
  authTime: Authentication time in milliseconds.
  dhcpTime: Time to obtain IP address in milliseconds.
  dnsTime: Time to resolve domain names in milliseconds.
type: object
properties:
  keys:
    mac:
      type: string
      description: Client MAC address
       pattern: "^([0-9A-F]{2}:){5}([0-9A-F]{2})$"
  ifIndex:
    description: radio interface.
    type: integer
    format: int32
  ssid:
    description: Client SSID name
    type: string
  txPackets:
    description: TX data frame count.
    type: integer
    format: int32
  txBytes:
    description: TX data byte count.
    type: integer
    format: int32
  txDrop:
    description: TX frames are dropped due to max retried.
    type: integer
    format: int32
  slaDrop:
    description: number of SAL violation traps occurred.
    type: integer
    format: int32
  rxPackets:
    description: RX data frame count.
    type: integer
    format: int32
  rxBytes:
    description: RX data byte count.
    type: integer
    format: int32
  rxDrop:
    description: RX frames dropped due to decrypt, MIC error etc.
    type: integer
    format: int32
  avgSnr:
    description: Average SNR (dB).
    type: integer
    format: int8
  psTimes:
    description: Client entered into power save mode times.
    type: integer
    format: int32
  radioScore:
    description: Radio link score (client SLA score).
    type: integer
    format: int8
  ipNetScore:
    description: IP network connectivity score.
    type: integer
    format: int8
  appScore:
    description: Application health score.
    type: integer
    format: int8
  phyMode:
    description: client radio mode.
    type: integer
    format: int8
  rssi:
    description: client rssi.
    type: integer
    format: int8
  os:
    description: client OS name
    type: string
  name:
    description: user name
    type: string
  host:  
    description: host name
    type: string
  profName:
    description: user profile name
    type: string
  dhcpIp:
    description: DHCP server IP.
    type: integer
    format: int32
  gwIp:
    description: gateway IP.
    type: integer
    format: int32
  dnsIp:
    description: DNS server IP.
    type: integer
    format: int32
  clientIp:
    description: Client IP address.
    type: integer
    format: int32
  dhcpTime:
    type: integer
    format: int32
  gwTime:
    type: integer
    format: int32
  dnsTime:
    type: integer
    format: int32
  clientTime:
    type: integer
    format: int64
  rxRateStats:
    description: rx rate counter
    type: array

  txRateStats:
    description: tx rate counter
    type: array

  txNssUsage:
    description: TX spatial stream percentage.
    type: array
    maxItems: 4
    items:
      type: integer
      format: int8
  txAirtime:
    description: station TX airtime percentage.

  rxAirtime:
    description: station RX airtime percentage.

  bwUsage:
    description: Bandwidth usages.

  required:
    keys

## Troubleshooting

On HiveOS, we can check output by executing hiden commands:
_show report snapshot client
and
_show report snapshot interface

## Example Output

```shell
{
  'rfStats': [
    {
      'device': {
        'serialnumber': '${SERIALNUM}'
      },
      'items': [
        {
          'bsSpCnt': 0,
          'channelUtilization': {
            'avg': 0,
            'max': 0,
            'min': 0
          },
          'clientCount': 1,
          'crcErrorRate': {
            'avg': 1437555618742272,
            'max': 1437555618742272,
            'min': 1437555618742272
          },
          'interferenceUtilization': {
            'avg': 0,
            'max': 0,
            'min': 0
          },
          'keys': {
            'ifindex': 17,
            'name': 'wifi0'
          },
          'lbSpCnt': 0,
          'noise': {
            'avg': 0,
            'max': 0,
            'min': 0
          },
          'rxAuthCnt': 0,
          'rxBcastBytes': 0,
          'rxBcastPackets': 0,
          'rxBitrateSuc': 0,
          'rxBytes': 243859054,
          'rxDropped': 0,
          'rxErrors': 0,
          'rxInbssUtilization': {
            'avg': 0,
            'max': 0,
            'min': 0
          },
          'rxMulticastBytes': 65442,
          'rxMulticastPackets': 0,
          'rxObssUtilization': {
            'avg': 0,
            'max': 0,
            'min': 0
          },
          'rxPackets': 623559,
          'rxPrbSpCnt': 0,
          'rxProbeSup': 0,
          'rxRate': {
            'avg': 0,
            'max': 0,
            'min': 0
          },
          'rxRateStats': {
            '0': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '1': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '10': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '100': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '101': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '102': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '103': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '104': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '105': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '106': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '107': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '108': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '109': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '11': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '110': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '111': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '112': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '113': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '114': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '115': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '116': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '117': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '118': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '119': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '12': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '120': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '121': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '122': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '123': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '124': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '125': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '126': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '127': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '128': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '129': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '13': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '130': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '131': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '132': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '133': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '134': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '135': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '136': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '137': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '138': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '139': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '14': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '140': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '141': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '142': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '143': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '144': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '145': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '146': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '147': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '148': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '149': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '15': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '150': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '151': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '152': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '153': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '154': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '155': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '156': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '157': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '158': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '159': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '16': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '160': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '161': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '162': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '163': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '164': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '165': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '166': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '167': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '168': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '169': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '17': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '170': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '171': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '172': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '173': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '174': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '175': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '176': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '177': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '178': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '179': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '18': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '180': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '181': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '182': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '183': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '184': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '185': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '186': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '187': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '188': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '189': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '19': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '190': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '191': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '2': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '20': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '21': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '22': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '23': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '24': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '25': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '26': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '27': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '28': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '29': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '3': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '30': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '31': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '32': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '33': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '34': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '35': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '36': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '37': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '38': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '39': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '4': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '40': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '41': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '42': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '43': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '44': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '45': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '46': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '47': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '48': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '49': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '5': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '50': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '51': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '52': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '53': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '54': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '55': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '56': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '57': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '58': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '59': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '6': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '60': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '61': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '62': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '63': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '64': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '65': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '66': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '67': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '68': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '69': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '7': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '70': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '71': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '72': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '73': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '74': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '75': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '76': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '77': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '78': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '79': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '8': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '80': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '81': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '82': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '83': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '84': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '85': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '86': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '87': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '88': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '89': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '9': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '90': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '91': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '92': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '93': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '94': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '95': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '96': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '97': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '98': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '99': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            }
          },
          'rxRetryCount': 0,
          'rxSwDropped': 0,
          'rxUnicastPackets': 0,
          'rxUtilization': {
            'avg': 0,
            'max': 0,
            'min': 0
          },
          'snAnswerCnt': 0,
          'snrSpCnt': 0,
          'txBcastBytes': 0,
          'txBcastPackets': 0,
          'txBitrateSuc': 0,
          'txBytes': 992704,
          'txDropped': 0,
          'txErrors': 697,
          'txHwDropped': 0,
          'txMulticastBytes': 0,
          'txMulticastPackets': 0,
          'txPackets': 7794,
          'txRate': {
            'avg': 0,
            'max': 0,
            'min': 0
          },
          'txRateStats': {
            '0': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '1': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '10': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '100': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '101': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '102': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '103': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '104': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '105': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '106': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '107': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '108': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '109': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '11': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '110': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '111': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '112': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '113': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '114': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '115': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '116': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '117': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '118': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '119': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '12': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '120': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '121': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '122': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '123': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '124': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '125': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '126': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '127': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '128': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '129': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '13': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '130': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '131': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '132': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '133': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '134': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '135': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '136': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '137': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '138': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '139': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '14': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '140': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '141': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '142': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '143': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '144': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '145': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '146': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '147': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '148': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '149': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '15': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '150': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '151': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '152': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '153': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '154': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '155': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '156': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '157': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '158': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '159': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '16': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '160': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '161': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '162': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '163': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '164': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '165': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '166': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '167': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '168': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '169': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '17': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '170': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '171': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '172': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '173': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '174': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '175': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '176': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '177': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '178': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '179': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '18': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '180': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '181': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '182': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '183': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '184': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '185': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '186': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '187': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '188': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '189': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '19': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '190': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '191': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '2': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '20': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '21': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '22': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '23': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '24': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '25': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '26': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '27': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '28': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '29': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '3': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '30': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '31': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '32': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '33': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '34': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '35': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '36': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '37': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '38': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '39': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '4': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '40': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '41': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '42': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '43': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '44': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '45': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '46': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '47': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '48': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '49': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '5': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '50': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '51': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '52': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '53': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '54': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '55': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '56': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '57': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '58': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '59': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '6': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '60': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '61': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '62': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '63': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '64': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '65': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '66': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '67': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '68': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '69': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '7': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '70': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '71': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '72': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '73': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '74': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '75': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '76': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '77': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '78': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '79': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '8': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '80': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '81': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '82': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '83': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '84': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '85': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '86': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '87': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '88': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '89': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '9': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '90': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '91': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '92': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '93': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '94': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '95': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '96': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '97': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '98': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '99': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            }
          },
          'txRetryCount': 0,
          'txSwDropped': 0,
          'txUnicastPackets': 0,
          'txUtilization': {
            'avg': 0,
            'max': 0,
            'min': 0
          },
          'wifinterferenceUtilization': {
            'avg': 0,
            'max': 0,
            'min': 0
          }
        },
        {
          'bsSpCnt': 0,
          'channelUtilization': {
            'avg': 0,
            'max': 0,
            'min': 0
          },
          'clientCount': 1,
          'crcErrorRate': {
            'avg': 789668392075264,
            'max': 789668392075264,
            'min': 789668392075264
          },
          'interferenceUtilization': {
            'avg': 0,
            'max': 0,
            'min': 0
          },
          'keys': {
            'ifindex': 15,
            'name': 'wifi1'
          },
          'lbSpCnt': 0,
          'noise': {
            'avg': 0,
            'max': 0,
            'min': 0
          },
          'rxAuthCnt': 0,
          'rxBcastBytes': 9746,
          'rxBcastPackets': 0,
          'rxBitrateSuc': 0,
          'rxBytes': 243860736,
          'rxDropped': 0,
          'rxErrors': 0,
          'rxInbssUtilization': {
            'avg': 0,
            'max': 0,
            'min': 0
          },
          'rxMulticastBytes': 65440,
          'rxMulticastPackets': 0,
          'rxObssUtilization': {
            'avg': 0,
            'max': 0,
            'min': 0
          },
          'rxPackets': 623563,
          'rxPrbSpCnt': 0,
          'rxProbeSup': 0,
          'rxRate': {
            'avg': 0,
            'max': 0,
            'min': 0
          },
          'rxRateStats': {
            '0': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '1': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '10': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '100': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '101': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '102': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '103': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '104': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '105': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '106': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '107': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '108': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '109': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '11': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '110': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '111': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '112': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '113': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '114': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '115': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '116': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '117': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '118': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '119': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '12': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '120': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '121': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '122': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '123': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '124': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '125': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '126': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '127': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '128': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '129': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '13': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '130': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '131': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '132': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '133': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '134': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '135': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '136': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '137': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '138': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '139': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '14': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '140': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '141': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '142': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '143': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '144': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '145': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '146': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '147': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '148': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '149': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '15': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '150': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '151': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '152': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '153': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '154': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '155': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '156': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '157': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '158': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '159': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '16': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '160': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '161': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '162': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '163': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '164': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '165': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '166': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '167': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '168': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '169': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '17': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '170': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '171': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '172': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '173': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '174': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '175': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '176': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '177': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '178': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '179': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '18': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '180': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '181': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '182': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '183': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '184': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '185': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '186': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '187': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '188': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '189': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '19': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '190': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '191': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '2': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '20': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '21': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '22': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '23': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '24': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '25': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '26': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '27': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '28': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '29': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '3': {
              'kbps': 1,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '30': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '31': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '32': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '33': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '34': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '35': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '36': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '37': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '38': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '39': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '4': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '40': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '41': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '42': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '43': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '44': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '45': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '46': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '47': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '48': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '49': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '5': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '50': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '51': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '52': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '53': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '54': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '55': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '56': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '57': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '58': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '59': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '6': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '60': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '61': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '62': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '63': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '64': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '65': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '66': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '67': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '68': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '69': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '7': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '70': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '71': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '72': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '73': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '74': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '75': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '76': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '77': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '78': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '79': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '8': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '80': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '81': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '82': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '83': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '84': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '85': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '86': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '87': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '88': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '89': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '9': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '90': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '91': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '92': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '93': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '94': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '95': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '96': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '97': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '98': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '99': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            }
          },
          'rxRetryCount': 210,
          'rxSwDropped': 0,
          'rxUnicastPackets': 0,
          'rxUtilization': {
            'avg': 0,
            'max': 0,
            'min': 0
          },
          'snAnswerCnt': 0,
          'snrSpCnt': 0,
          'txBcastBytes': 214994,
          'txBcastPackets': 0,
          'txBitrateSuc': 0,
          'txBytes': 992704,
          'txDropped': 0,
          'txErrors': 697,
          'txHwDropped': 697,
          'txMulticastBytes': 776836,
          'txMulticastPackets': 0,
          'txPackets': 7794,
          'txRate': {
            'avg': 0,
            'max': 0,
            'min': 0
          },
          'txRateStats': {
            '0': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '1': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '10': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '100': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '101': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '102': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '103': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '104': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '105': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '106': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '107': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '108': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '109': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '11': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '110': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '111': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '112': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '113': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '114': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '115': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '116': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '117': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '118': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '119': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '12': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '120': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '121': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '122': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '123': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '124': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '125': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '126': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '127': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '128': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '129': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '13': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '130': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '131': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '132': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '133': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '134': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '135': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '136': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '137': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '138': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '139': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '14': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '140': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '141': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '142': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '143': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '144': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '145': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '146': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '147': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '148': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '149': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '15': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '150': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '151': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '152': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '153': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '154': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '155': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '156': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '157': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '158': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '159': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '16': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '160': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '161': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '162': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '163': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '164': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '165': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '166': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '167': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '168': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '169': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '17': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '170': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '171': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '172': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '173': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '174': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '175': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '176': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '177': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '178': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '179': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '18': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '180': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '181': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '182': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '183': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '184': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '185': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '186': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '187': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '188': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '189': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '19': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '190': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '191': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '2': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '20': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '21': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '22': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '23': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '24': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '25': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '26': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '27': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '28': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '29': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '3': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '30': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '31': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '32': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '33': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '34': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '35': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '36': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '37': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '38': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '39': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '4': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '40': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '41': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '42': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '43': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '44': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '45': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '46': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '47': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '48': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '49': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '5': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '50': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '51': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '52': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '53': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '54': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '55': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '56': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '57': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '58': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '59': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '6': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '60': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '61': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '62': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '63': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '64': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '65': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '66': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '67': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '68': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '69': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '7': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '70': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '71': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '72': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '73': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '74': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '75': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '76': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '77': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '78': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '79': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '8': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '80': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '81': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '82': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '83': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '84': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '85': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '86': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '87': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '88': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '89': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '9': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '90': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '91': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '92': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '93': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '94': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '95': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '96': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '97': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '98': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            },
            '99': {
              'kbps': 0,
              'rateDtn': 0,
              'rateSucDtn': 0
            }
          },
          'txRetryCount': 0,
          'txSwDropped': 0,
          'txUnicastPackets': 0,
          'txUtilization': {
            'avg': 0,
            'max': 0,
            'min': 0
          },
          'wifinterferenceUtilization': {
            'avg': 0,
            'max': 0,
            'min': 0
          }
        }
      ],
      'name': 'RfStats',
      'ts': 1722327710
    }
  ]
}

```
