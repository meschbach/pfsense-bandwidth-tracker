package iftop

import "fmt"

type ByteReading string

const ScaleKB = 1024
const ScaleMB = ScaleKB * 1024
const ScaleGB = ScaleMB * 1024
const ScaleTB = ScaleGB * 1024

func (b ByteReading) ToFloat64() float64 {
	var value float64
	var suffix string
	fmt.Sscanf(string(b), "%f%s", &value, &suffix)
	scale := 1.0
	switch suffix {
	case "B":
		break
	case "KB":
		scale = ScaleKB
	case "MB":
		scale = ScaleMB
	case "GB":
		scale = ScaleGB
	case "TB":
		scale = ScaleTB
	}
	return value * scale
}
