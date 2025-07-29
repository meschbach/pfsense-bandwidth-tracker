package netstat

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

const simpleCase = "Name       Mtu Network             Address                               Ipkts Ierrs Idrop         Ibytes       Opkts Oerrs         Obytes  Coll  Drop\nigb0      1500 <Link#1>            00:90:0b:7c:06:00                1455471621     2     0  1773036112075   473397831    11   124176101655     3     7 \nigb0         - fe80::%igb0/64      fe80::290:bff:fe7c:602%igb0               1     -     -              3           2     -            156     -     - "

func TestSimpleCase(t *testing.T) {
	i := &interpreter{state: StateStart}
	for _, line := range strings.Split(simpleCase, "\n") {
		require.NoError(t, i.consumeLine(line))
	}
	result, err := i.done()
	require.NoError(t, err)
	if assert.Len(t, result, 1) {
		assert.Equal(t, "igb0", result[0].Name)
		assert.Equal(t, 1500, result[0].MTU)
		assert.Equal(t, int64(2), result[0].IngressErrors)
		assert.Equal(t, int64(0), result[0].IngressDrop)
		assert.Equal(t, int64(1455471621), result[0].IfaceStats.Ingress.Packets)
		assert.Equal(t, int64(1773036112075), result[0].IfaceStats.Ingress.Bytes)

		assert.Equal(t, int64(11), result[0].EgressErrors)
		assert.Equal(t, int64(7), result[0].Drop)
		assert.Equal(t, int64(3), result[0].Collisions)
		assert.Equal(t, int64(473397831), result[0].IfaceStats.Egress.Packets, "egress packets wrong")
		assert.Equal(t, int64(124176101655), result[0].IfaceStats.Egress.Bytes, "egress bytes wrong")

		if assert.Len(t, result[0].AddressReadings, 1) {
			assert.Equal(t, "fe80::%igb0/64", result[0].AddressReadings[0].Network)
			assert.Equal(t, "fe80::290:bff:fe7c:602%igb0", result[0].AddressReadings[0].Address)
			assert.Equal(t, int64(1), result[0].AddressReadings[0].Reading.Ingress.Packets)
			assert.Equal(t, int64(3), result[0].AddressReadings[0].Reading.Ingress.Bytes)
			assert.Equal(t, int64(2), result[0].AddressReadings[0].Reading.Egress.Packets)
			assert.Equal(t, int64(156), result[0].AddressReadings[0].Reading.Egress.Bytes)
		}
	}
}
