package systemd

import (
	"testing"

	"github.com/coreos/go-systemd/dbus"
	godbus "github.com/godbus/dbus"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockDbusConn struct {
	mock.Mock
}

func (m MockDbusConn) GetUnitProperty(unit string, propertyName string) (*dbus.Property, error) {
	args := m.Called(unit, propertyName)
	return args.Get(0).(*dbus.Property), args.Error(1)
}

func (m MockDbusConn) GetUnitTypeProperty(unit string, unitType string, propertyName string) (*dbus.Property, error) {
	args := m.Called(unit, unitType, propertyName)
	return args.Get(0).(*dbus.Property), args.Error(1)
}

func mockDbusProperty(value interface{}) *dbus.Property {
	variant := godbus.MakeVariant(value)
	return &dbus.Property{
		Value: variant,
	}
}

func TestCollectActiveState(t *testing.T) {
	unitName := "testUnit"
	activeEnterTimestamp := uint64(1234)

	unit := dbus.UnitStatus{
		ActiveState: "active",
		Name:        unitName,
	}

	conn := new(MockDbusConn)
	conn.On("GetUnitProperty", unitName, "ActiveEnterTimestamp").Return(mockDbusProperty(activeEnterTimestamp), nil)

	fields := map[string]interface{}{}

	collectActiveState(unit, conn, fields)

	require.Equal(t, 1, fields["isActive"])
	require.Equal(t, activeEnterTimestamp, fields["activeEnterTimestamp"])
}

func TestCollectPerUnitType(t *testing.T) {
	unitName := "testUnit"
	lastTriggerValue := uint64(1234)
	nRestarts := uint32(2345)
	nAccepted := uint32(3456)
	nConnection := uint32(4567)
	nRefused := uint32(5678)

	unit := dbus.UnitStatus{
		Name: unitName,
	}

	conn := new(MockDbusConn)
	conn.On("GetUnitTypeProperty", unitName, "Timer", "LastTriggerUSec").Return(mockDbusProperty(lastTriggerValue), nil)
	conn.On("GetUnitTypeProperty", unitName, "Service", "NRestarts").Return(mockDbusProperty(nRestarts), nil)
	conn.On("GetUnitTypeProperty", unitName, "Socket", "NAccepted").Return(mockDbusProperty(nAccepted), nil)
	conn.On("GetUnitTypeProperty", unitName, "Socket", "NConnection").Return(mockDbusProperty(nConnection), nil)
	conn.On("GetUnitTypeProperty", unitName, "Socket", "NRefused").Return(mockDbusProperty(nRefused), nil)

	tags := map[string]string{}
	fields := map[string]interface{}{}
	collectTimerUnit(unit, conn, tags, fields)
	require.Equal(t, "Timer", tags["unitType"])
	require.Equal(t, lastTriggerValue, fields["lastTriggerUSec"])

	tags = map[string]string{}
	fields = map[string]interface{}{}
	collectServiceUnit(unit, conn, tags, fields)
	require.Equal(t, "Service", tags["unitType"])
	require.Equal(t, nRestarts, fields["nRestarts"])

	tags = map[string]string{}
	fields = map[string]interface{}{}
	collectSocketUnit(unit, conn, tags, fields)
	require.Equal(t, "Socket", tags["unitType"])
	require.Equal(t, nAccepted, fields["nAccepted"])
	require.Equal(t, nConnection, fields["nConnection"])
	require.Equal(t, nRefused, fields["nRefused"])
}
