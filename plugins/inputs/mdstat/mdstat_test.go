package mdstat

import (
	"strings"
	"testing"

	//"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

func TestValidPersonalities(t *testing.T) {
	exampleString := "Personalities : [raid1] [raid6]"

	result, _ := parsePersonalities(exampleString)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, "[raid1]", result[0])
	assert.Equal(t, "[raid6]", result[1])

	exampleString = "Personalities : [raid1]"
	result, _ = parsePersonalities(exampleString)
	assert.Equal(t, 1, len(result))

	exampleString = "Personalities : "
	result, _ = parsePersonalities(exampleString)
	assert.Equal(t, 0, len(result))
}

func TestBasicDeviceParsing(t *testing.T) {
	exampleFile := `Personalities : [raid1] [raid6] [raid5] [raid4]
md_d0 : active raid5 sde1[0] sdf1[4] sdb1[5] sdd1[2] sdc1[1]
      1250241792 blocks super 1.2 level 5, 64k chunk, algorithm 2 [5/5] [UUUUU]
      bitmap: 0/10 pages [0KB], 16384KB chunk

unused devices: <none>`

	expectedD1 := disk{"sde1", 0, false}
	expectedD2 := disk{"sdf1", 4, false}
	expectedD3 := disk{"sdb1", 5, false}
	expectedD4 := disk{"sdd1", 2, false}
	expectedD5 := disk{"sdc1", 1, false}

	result, _ := parseFile(strings.NewReader(exampleFile))
	device := result.devices[0]

	assert.Equal(t, "md_d0", device.name)
	assert.Equal(t, "active", device.status)
	assert.Equal(t, "raid5", device.raidType)

	assert.Equal(t, 5, len(device.diskList))
	assert.Equal(t, expectedD1, device.diskList[0])
	assert.Equal(t, expectedD2, device.diskList[1])
	assert.Equal(t, expectedD3, device.diskList[2])
	assert.Equal(t, expectedD4, device.diskList[3])
	assert.Equal(t, expectedD5, device.diskList[4])

	assert.Equal(t, 5, device.minDisks)
	assert.Equal(t, 5, device.currDisks)
	assert.Equal(t, 0, device.missingDisks)
	assert.Equal(t, 0, device.failedDisks)
}

func TestFailedDisk(t *testing.T) {
	exampleFile := `Personalities : [raid1]
md1 : active raid1 sde1[6](F) sdg1[1] sdb1[4](F) sdd1[3] sdc1[2]
      488383936 blocks [6/4] [_UUU_U]

unused devices: <none>`

	result, _ := parseFile(strings.NewReader(exampleFile))
	device := result.devices[0]

	assert.Equal(t, true, device.diskList[0].failed)
	assert.Equal(t, false, device.diskList[1].failed)
	assert.Equal(t, true, device.diskList[2].failed)
	assert.Equal(t, 2, device.failedDisks)
	assert.Equal(t, 6, device.minDisks)
	assert.Equal(t, 4, device.currDisks)
	assert.Equal(t, 0, device.missingDisks)
}

func TestMissingDisks(t *testing.T) {
	exampleFile := `Personalities : [raid1]
md1 : active raid1 sde1[6](F) sdg1[1] sdb1[4] sdd1[3] sdc1[2]
      488383936 blocks [6/4] [_UUUU_]

unused devices: <none>`

	result, _ := parseFile(strings.NewReader(exampleFile))
	device := result.devices[0]

	assert.Equal(t, 1, device.failedDisks)
	assert.Equal(t, 6, device.minDisks)
	assert.Equal(t, 4, device.currDisks)
	assert.Equal(t, 1, device.missingDisks)
}

func TestBasicParsing(t *testing.T) {
	exampleFile := `Personalities : [raid1] [raid6] [raid5] [raid4]
md_d0 : active raid5 sde1[0] sdf1[4] sdb1[5] sdd1[2] sdc1[1]
      1250241792 blocks super 1.2 level 5, 64k chunk, algorithm 2 [5/5] [UUUUU]
      bitmap: 0/10 pages [0KB], 16384KB chunk

unused devices: <none>
  `

	result, _ := parseFile(strings.NewReader(exampleFile))
	assert.Equal(t, 4, len(result.personalities))
	assert.Equal(t, 1, len(result.devices))
	assert.Equal(t, "md_d0", result.devices[0].name)

}

func TestMultiDeviceParsing(t *testing.T) {
	exampleFile := `Personalities : [raid1] [raid6] [raid5] [raid4]
md1 : active raid1 sdb2[1] sda2[0]
      136448 blocks [2/2] [UU]

md2 : active raid1 sdb3[1] sda3[0]
      129596288 blocks [2/2] [UU]

md3 : active raid5 sdl1[9] sdk1[8] sdj1[7] sdi1[6] sdh1[5] sdg1[4] sdf1[3] sde1[2] sdd1[1] sdc1[0]
      1318680576 blocks level 5, 1024k chunk, algorithm 2 [10/10] [UUUUUUUUUU]

md0 : active raid1 sdb1[1] sda1[0]
      16787776 blocks [2/2] [UU]

unused devices: <none>`

	result, _ := parseFile(strings.NewReader(exampleFile))
	assert.Equal(t, 4, len(result.devices))
	assert.Equal(t, "md1", result.devices[0].name)
	assert.Equal(t, "md2", result.devices[1].name)
	assert.Equal(t, "md3", result.devices[2].name)
	assert.Equal(t, "md0", result.devices[3].name)

}

func TestRecoveryParsing(t *testing.T) {
	exampleFile := `Personalities : [raid1] [raid6] [raid5] [raid4]
md0 : active raid5 sdh1[6] sdg1[4] sdf1[3] sde1[2] sdd1[1] sdc1[0]
      1464725760 blocks level 5, 64k chunk, algorithm 2 [6/5] [UUUUU_]
      [==>..................]  recovery = 12.6% (37043392/292945152) finish=127.5min speed=33440K/sec

`

	result, _ := parseFile(strings.NewReader(exampleFile))
	//result := parseFile(strings.NewReader(exampleFile))
	device := result.devices[0]
	assert.True(t, device.inRecovery)
	assert.InDelta(t, 12.6, device.recoveryPercent, 0.001)
}
