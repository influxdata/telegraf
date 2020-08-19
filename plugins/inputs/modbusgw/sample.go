package modbusgw

const sampleConfig = `
[[inputs.modbusgw]]
    #
    # Name of this input - should be unique
    #
    name="sma"

    #
    # Address and port of the modbus server or gateway
    #
    gateway="tcp://yourserver.com:502"

    #
    # Response timeout, in go duration fornat
    # Usually can be set pretty low
    #
    timeout="5s"

    #
    # Request (poll) definitions
    #
    # Request parameters:
    #
    # unit - required.  Unit address of device being polled.  Per spec, the value is between
    #    1 and 247, or 0 for broadcast.  The values 0 or 255 are usually accepted to communicate
    #    directly to a Modbus/TCP device not acting as a gateway.  Historically this has been a
    #    point of confusion.  If this is what you want (to talk to the gateway itself), try 255 first.
    #    Using the broadcast address can cause unexpected device responses.
    #
    # address - the register address of the first register being requested.  This address is zero-based.
    #    For example, the first holding register is address 0.  Be aware that some documentation
    #
    # count - how mant 16-bit registers to request
    #
    # type - defines the register type, which maps internally to the function code used ub the
    #   PDU (request).  Must be "holding" or "input", if unspecified defaults to "holding"
    #
    # measurement - the nameof the measurement, for example when stored in influx
    #
    # fields - defines how the response PDU is mapped to fields of the measurement.  Attributes
    # of each field are:
    #
    # name - name of the field
    #
    # type - must be INT32, UINT32, INT16, or UINT16.  More tyoes will be added in the future.
    #
    # scale, offset - math performed on the raw modbus value before storing.
    #    stored field value = (modbus value * scale) + offset
    #
    # omit - if true, don't store this field at all.  you must still set a 'type'.  Use this to
    #   skip fields not of interest that are part of the response because they are within the
    #   requested register range.
    #
    requests = [
        { unit=3, address=30769, count=8, type="holding", measurement="pv1", fields = [
            {name="Ipv", type="INT32", scale=0.001},
            {name="Vpv", type="INT32", scale=0.01},
            {name="Ppv", type="INT32", omit=true},
            {name="Pac", type="INT32", scale=1.0},
        ] },
        { unit=4, address=30769, count=8, type="holding", measurement="pv2", fields = [
            {name="Ipv", type="INT32", scale=0.001},
            {name="Vpv", type="INT32", scale=0.01},
            {name="Ppv", type="INT32", omit=true},
            {name="Pac", type="INT32", scale=1.0},
        ] },
        { unit=5, address=30769, count=8, type="holding", measurement="pv3", fields = [
            {name="Ipv", type="INT32", scale=0.001},
            {name="Vpv", type="INT32", scale=0.01},
            {name="Ppv", type="INT32", omit=true},
            {name="Pac", type="INT32", scale=1.0},
        ] },
        { unit=6, address=30769, count=8, type="holding", measurement="pv4", fields = [
            {name="Ipv", type="INT32", scale=0.001},
            {name="Vpv", type="INT32", scale=0.01},
            {name="Ppv", type="INT32", omit=true},
            {name="Pac", type="INT32", scale=1.0},
        ] },
        { unit=7, address=30769, count=8, type="holding", measurement="pv5", fields = [
            {name="Ipv", type="INT32", scale=0.001},
            {name="Vpv", type="INT32", scale=0.01},
            {name="Ppv", type="INT32", omit=true},
            {name="Pac", type="INT32", scale=1.0},
        ] },
        { unit=8, address=30769, count=8, type="holding", measurement="pv6", fields = [
            {name="Ipv", type="INT16", scale=0.001},
            {name="Vpv", type="INT16", scale=0.01},
            {name="Ppv", type="INT32", omit=true},
            {name="Pac", type="INT32", scale=1.0},
        ] },
        { unit=9, address=30769, count=8, type="holding", measurement="pv7", fields = [
            {name="Ipv", type="INT32", scale=0.001},
            {name="Vpv", type="INT32", scale=0.01},
            {name="Ppv", type="INT32", omit=true},
            {name="Pac", type="INT32", scale=1.0},
        ] },
     ]
`

func (m *ModbusGateway) SampleConfig() string {
	return sampleConfig
}
