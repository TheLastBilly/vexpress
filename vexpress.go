package main

import (
	"fmt"
	"math"

	"encoding/binary"
)

const OSF_FRAME_SIZE int	= 1785

type OSFRawFrame			[OSF_FRAME_SIZE]byte

type Vector2 [2]float32
type Vector3 [3]float32

type OSFFrame struct {
	Now						float64

	Id						int32

	Width					float32
	Height					float32

	EyeBlinkRight			float32
	EyeBlinkLeft			float32

	Success					byte

	PNPError				float32
	
	Quaternion				[4]float32
	Euler					[3]float32
	Translation				[3]float32

	LMSConfidence			[68]float32
	LMS						[68]Vector2

	PNPPoints				[70]Vector3

	EyeLeft					float32
	EyeRight				float32

	EyeSteepnessLeft		float32
	EveUpDownLeft			float32
	EyeQuirkLeft			float32
	
	EyeSteepnessRight		float32
	EveUpDownRight			float32
	EyeQuirkRight			float32

	MouthCornerUpDownLeft	float32
	MouthCornerInOutLeft	float32
	MouthCornerUpDownRight	float32
	MouthCornerInOutRight	float32

	MouthOpen				float32
	MouthWide				float32
}

func float64FromBytes(b []byte) float64 {
	return math.Float64frombits(binary.LittleEndian.Uint64(b))
}
func float32FromBytes(b []byte) float32 {
	return math.Float32frombits(binary.LittleEndian.Uint32(b))
}
func int32FromBytes(b []byte) int32 {
	var v int32 = 0

	v |= int32(b[0])
	v |= int32(b[1]) << 8
	v |= int32(b[2]) << 16
	v |= int32(b[3]) << 24

	return v
}

func OSFParseFrame(raw *OSFRawFrame) (*OSFFrame, error) {
	var i int32 = 0
	var e int32 = 0
	var s int32 = 0
	var err error = nil
	frame := &OSFFrame{}

	fetchByte := func() byte {
		v := raw[i]; i++
		return v
	}
	fetchInt32 := func() int32 {
		v := int32FromBytes(raw[i:i+3]); i+=4
		return v
	}
	fetchFloat32 := func() float32 {
		v := float32FromBytes(raw[i:i+3]); i+=4
		return v
	}
	fetchFloat64 := func() float64 {
		v := float64FromBytes(raw[i:i+7]); i+=8
		return v
	}

	frame.Now = fetchFloat64()

	frame.Id = fetchInt32()
	
	frame.Width = fetchFloat32()
	frame.Height = fetchFloat32()

	frame.EyeBlinkRight = fetchFloat32()
	frame.EyeBlinkLeft = fetchFloat32()

	frame.Success = fetchByte()

	frame.PNPError = fetchFloat32()

	s = 0
	e = i+(4*4)
	for ; i < e; s++ {
		frame.Quaternion[s] = fetchFloat32()
	}

	s = 0
	e = i+(4*3)
	for ; i < e; s++ {
		frame.Euler[s] = fetchFloat32()
	}

	s = 0
	e = i+(4*3)
	for ; i < e; s++ {
		frame.Translation[s] = fetchFloat32()
	}

	return nil, err
}

func main() {
	fmt.Print("Hello World\n")
}