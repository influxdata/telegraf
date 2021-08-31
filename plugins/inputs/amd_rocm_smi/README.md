# ROCm System Management Interface (SMI) Input Plugin

This plugin uses a query on the [`rocm-smi`](https://github.com/RadeonOpenCompute/rocm_smi_lib/tree/master/python_smi_tools) binary to pull GPU stats including memory and GPU usage, temperatures and other.

### Configuration

```toml
# Pulls statistics from nvidia GPUs attached to the host
[[inputs.amd_rocm_smi]]
  ## Optional: path to rocm-smi binary, defaults to $PATH via exec.LookPath
  # bin_path = "/opt/rocm/bin/rocm-smi"

  ## Optional: timeout for GPU polling
  # timeout = "5s"
```

### Metrics
- measurement: `amd_rocm_smi`
  - tags
    - `name` (entry name assigned by rocm-smi executable)
    - `gpu_id` (id of the GPU according to rocm-smi)
    - `gpu_unique_id` (unique id of the GPU)

  - fields
    - `driver_version` (integer)
    - `fan_speed`(integer)
    - `memory_total`(integer B)
    - `memory_used`(integer B)
    - `memory_free`(integer B)
    - `temperature_sensor_edge` (float, Celsius)
    - `temperature_sensor_junction` (float, Celsius)
    - `temperature_sensor_memory` (float, Celsius)
    - `utilization_gpu` (integer, percentage)
    - `utilization_memory` (integer, percentage)
    - `clocks_current_sm` (integer, Mhz)
    - `clocks_current_memory` (integer, Mhz)
    - `power_draw` (float, Watt)

### Troubleshooting
Check the full output by running `rocm-smi` binary manually.

Linux:
```sh
rocm-smi rocm-smi -o -l -m -M  -g -c -t -u -i -f -p -P -s -S -v --showreplaycount --showpids --showdriverversion --showmemvendor --showfwinfo --showproductname --showserial --showuniqueid --showbus --showpendingpages --showpagesinfo --showretiredpages --showunreservablepages --showmemuse --showvoltage --showtopo --showtopoweight --showtopohops --showtopotype --showtoponuma --showmeminfo all --json
```
Please include the output of this command if opening a GitHub issue, together with ROCm version.

### Example Output
```
{"card0": {"GPU ID": "0x6861", "Unique ID": "0x2150e7d042a1124", "VBIOS version": "113-D0510100-106", "Temperature (Sensor edge) (C)": "40.0", "Temperature (Sensor junction) (C)": "41.0", "Temperature (Sensor memory) (C)": "92.0", "dcefclk clock speed:": "(600Mhz)", "dcefclk clock level:": "0", "mclk clock speed:": "(945Mhz)", "mclk clock level:": "3", "sclk clock speed:": "(1269Mhz)", "sclk clock level:": "3", "socclk clock speed:": "(960Mhz)", "socclk clock level:": "3", "pcie clock level": "1 (8.0GT/s x16)", "sclk clock level": "3 (1269Mhz)", "Fan speed (level)": "33", "Fan speed (%)": "13", "Fan RPM": "680", "Performance Level": "auto", "GPU OverDrive value (%)": "0", "GPU Memory OverDrive value (%)": "0", "Max Graphics Package Power (W)": "170.0", "Average Graphics Package Power (W)": "15.0", "0": "8.0GT/s x16", "1": "8.0GT/s x16 *", "2": "847Mhz", "3": "960Mhz *", "4": "1028Mhz", "5": "1107Mhz", "6": "1440Mhz", "7": "1500Mhz", "GPU use (%)": "0", "GPU memory vendor": "samsung", "PCIe Replay Count": "0", "Serial Number": "N/A", "Voltage (mV)": "906", "PCI Bus": "0000:04:00.0", "VRAM Total Memory (B)": "17163091968", "VRAM Total Used Memory (B)": "17776640", "VIS_VRAM Total Memory (B)": "268435456", "VIS_VRAM Total Used Memory (B)": "13557760", "GTT Total Memory (B)": "17163091968", "GTT Total Used Memory (B)": "25608192", "ASD firmware version": "553648152", "CE firmware version": "79", "DMCU firmware version": "0", "MC firmware version": "0", "ME firmware version": "163", "MEC firmware version": "432", "MEC2 firmware version": "432", "PFP firmware version": "186", "RLC firmware version": "93", "RLC SRLC firmware version": "0", "RLC SRLG firmware version": "0", "RLC SRLS firmware version": "0", "SDMA firmware version": "430", "SDMA2 firmware version": "430", "SMC firmware version": "00.28.54.00", "SOS firmware version": "0x0008015d", "TA RAS firmware version": "00.00.00.00", "TA XGMI firmware version": "00.00.00.00", "UVD firmware version": "0x422b1100", "VCE firmware version": "0x39060400", "VCN firmware version": "0x00000000", "Card model": "0xc1e", "Card vendor": "Advanced Micro Devices, Inc. [AMD/ATI]", "Card SKU": "D05101", "(Topology) Numa Node": "0", "(Topology) Numa Affinity": "0"}, "system": {"Driver version": "5.9.25"}}

```

### Limitations and notices
Please notice that this plugin has been developed and tested on a limited number of versions and small set of GPUs. Currently the latest ROCm version tested is 4.3.0.
Notice that depending on the device and driver versions the amount of information provided by `rocm-smi` can vary so that some fields would start/stop appearing in the metrics upon updates.
The `rocm-smi` JSON output is not perfectly homogeneous and is possibly changing in the future, hence parsing and unmarshaling can start failing upon updating ROCm.

Inspired by the current state of the art of the `nvidia-smi` plugin.
