package netstat

import "fmt"

const KiloBytes int64 = 1024
const MegaBytes = 1024 * KiloBytes
const GigaBytes = 1024 * MegaBytes
const TeraBytes = 1024 * GigaBytes

type ScaledBytes struct {
	Placeholder string
	Value       int64
}

func (s ScaledBytes) String() string {
	return fmt.Sprintf("%5d %3s", s.Value, s.Placeholder)
}

// bytesToScaled will find the largest descriable unit and summarize
func bytesToScaled(value int64) ScaledBytes {
	if value < (KiloBytes * 10) {
		return ScaledBytes{Placeholder: "B", Value: value}
	} else if value < (MegaBytes * 10) {
		return ScaledBytes{Placeholder: "KiB", Value: value / KiloBytes}
	} else if value < (GigaBytes * 10) {
		return ScaledBytes{Placeholder: "MiB", Value: value / MegaBytes}
	} else if value < (TeraBytes * 10) {
		return ScaledBytes{Placeholder: "GiB", Value: value / GigaBytes}
	} else {
		return ScaledBytes{Placeholder: "TiB", Value: value / TeraBytes}
	}
}

const Thousands = 1000
const Millions = 1000 * Thousands
const Billions = 1000 * Millions
const Trillions = 1000 * Billions

func numberToScaled(value int64) ScaledBytes {
	if value < (Thousands * 10) {
		return ScaledBytes{Placeholder: "", Value: value}
	} else if value < (Millions * 10) {
		return ScaledBytes{Placeholder: "K", Value: value / Thousands}
	} else if value < (Billions * 10) {
		return ScaledBytes{Placeholder: "M", Value: value / Millions}
	} else if value < (Trillions * 10) {
		return ScaledBytes{Placeholder: "B", Value: value / Billions}
	} else {
		return ScaledBytes{Placeholder: "T", Value: value / Trillions}
	}
}
