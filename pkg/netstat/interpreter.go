package netstat

import (
	"fmt"
)

type InterpreterState int

const (
	StateStart InterpreterState = iota
	IfaceLine
)

type DirectionReading struct {
	Packets int64
	Bytes   int64
}

type Reading struct {
	Ingress DirectionReading
	Egress  DirectionReading
}

type AddressReading struct {
	Network string
	Address string
	Reading Reading
}

type IFaceReading struct {
	Name            string
	MTU             int
	IfaceStats      Reading
	IngressErrors   int64
	IngressDrop     int64
	EgressErrors    int64
	Collisions      int64
	Drop            int64
	AddressReadings []*AddressReading
}

type interpreter struct {
	state        InterpreterState
	readings     []*IFaceReading
	currentIFace *IFaceReading
}

// igb0      1500 <Link#1>            00:90:0b:7c:06:00                1455574454     2     0  1773163896201   473431445    11   124181250793     3     7
// igb0         - fe80::%igb0/64      fe80::290:bff:fe7c:602%igb0               0     -     -              0           2     -            156     -     -
func (i *interpreter) consumeLine(line string) error {
	switch i.state {
	case StateStart:
		i.state = IfaceLine
	case IfaceLine:
		var name, mtu, network, address, ierrs, idrop, oerrs, coll, drop string
		var ipkts, ibytes, opkts, obytes int64
		count, err := fmt.Sscanf(line, "%s\t%s\t%s\t%s\t%d\t%s\t%s\t%d\t%d\t%s\t%d\t%s\t%s", &name, &mtu, &network, &address, &ipkts, &ierrs, &idrop, &ibytes, &opkts, &oerrs, &obytes, &coll, &drop)
		if err != nil {
			return err
		}
		if count != 13 {
			return fmt.Errorf("expected 13 arguments, got %d", count)
		}
		if mtu == "-" {
			i.currentIFace.AddressReadings = append(i.currentIFace.AddressReadings, &AddressReading{
				Network: network,
				Address: address,
				Reading: Reading{
					Ingress: DirectionReading{
						Packets: ipkts,
						Bytes:   ibytes,
					},
					Egress: DirectionReading{
						Packets: opkts,
						Bytes:   obytes,
					},
				},
			})
		} else {
			i.currentIFace = &IFaceReading{
				Name: name,
				IfaceStats: Reading{
					Ingress: DirectionReading{
						Packets: ipkts,
						Bytes:   ibytes,
					},
					Egress: DirectionReading{
						Packets: opkts,
						Bytes:   obytes,
					},
				},
				AddressReadings: nil,
			}
			fmt.Sscanf(mtu, "%d", &i.currentIFace.MTU)
			fmt.Sscanf(ierrs, "%d", &i.currentIFace.IngressErrors)
			fmt.Sscanf(idrop, "%d", &i.currentIFace.IngressDrop)
			fmt.Sscanf(oerrs, "%d", &i.currentIFace.EgressErrors)
			fmt.Sscanf(coll, "%d", &i.currentIFace.Collisions)
			fmt.Sscanf(drop, "%d", &i.currentIFace.Drop)
			i.readings = append(i.readings, i.currentIFace)
		}
		i.state = IfaceLine
	default:
		panic("Unexpected state")
	}
	return nil
}

func (i *interpreter) done() ([]*IFaceReading, error) {
	return i.readings, nil
}
