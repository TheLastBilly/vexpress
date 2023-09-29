package main

import (
	"fmt"
	"net"
	"bytes"

	"math/rand"
	"encoding/binary"

	"libsvm-go"
)

const OSF_FRAME_SIZE		int 	= 1785
const OSF_DEFAULT_HOST 		string 	= "127.0.0.1"
const OSF_DEFAULT_PORT 		int 	= 11573

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

func (f OSFFrame) TrainingSlice() []float64 {
	i := 0
	s := make([]float64, 446)

	s[i] = float64(f.EyeBlinkRight); i++
	s[i] = float64(f.EyeBlinkLeft); i++

	for _, v := range f.Quaternion {
		s[i] = float64(v); i++
	}
	for _, v := range f.Euler {
		s[i] = float64(v); i++
	}
	for _, v := range f.Translation {
		s[i] = float64(v); i++
	}

	s[i] = float64(f.EyeLeft); i++
	s[i] = float64(f.EyeRight); i++
	s[i] = float64(f.EyeSteepnessLeft); i++
	s[i] = float64(f.EveUpDownLeft); i++
	s[i] = float64(f.EyeQuirkLeft); i++
	s[i] = float64(f.EyeSteepnessRight); i++
	s[i] = float64(f.EveUpDownRight); i++
	s[i] = float64(f.EyeQuirkRight); i++
	s[i] = float64(f.MouthCornerUpDownLeft); i++
	s[i] = float64(f.MouthCornerInOutLeft); i++
	s[i] = float64(f.MouthCornerUpDownRight); i++
	s[i] = float64(f.MouthCornerInOutRight); i++
	s[i] = float64(f.MouthOpen); i++
	s[i] = float64(f.MouthWide); i++

	return s
}

func OSFParseFrame(raw []byte) (*OSFFrame, error) {
	var err error = nil
	frame := &OSFFrame{}
	buf := bytes.NewBuffer(make([]byte, 0, OSF_FRAME_SIZE))

	if err = binary.Write(buf, binary.LittleEndian, raw); err != nil {
		return nil, err
	}

	if err = binary.Read(buf, binary.LittleEndian, frame); err != nil {
		return nil, err
	}

	return frame, err
}

func TrainExpressionsModel(attributes []libSvm.Attributes) (*libSvm.Model, error) {
	param := libSvm.NewParameter()
	param.KernelType = libSvm.POLY
	model := libSvm.NewModel(param)

	problem, err := libSvm.NewProblemFromAttributes(attributes, param) 
		
	model.Train(problem)

	return model, err
}

func BuildAttributeList(class float64, frames []OSFFrame) []libSvm.Attributes {
	attributes := make([]libSvm.Attributes, 0, len(frames))
	for _, frame := range frames {
		slice := frame.TrainingSlice()

		attr := libSvm.Attributes{
			Class : class,
			Snodes : make([]libSvm.Snode, 0, len(slice)),
		}

		for i, v := range slice {
			attr.Snodes = append(attr.Snodes, libSvm.Snode{
				Index: i + 1,
				Value: v,
			})
		}
		
		attributes = append(attributes, attr)
	}

	return attributes
}

func main() {
	buf := make([]byte, OSF_FRAME_SIZE)
	host := OSF_DEFAULT_HOST
	port := OSF_DEFAULT_PORT
	class := float64(1.0)
	attributes := make([]libSvm.Attributes, 0, 1000)

	uri := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.ListenPacket("udp", uri)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	fmt.Printf("waiting for OpenSeeFace\n")
	for class < 4.0 {
		frames := make([]OSFFrame, 0, 1000)
		for true {
			_, _, err := conn.ReadFrom(buf)
			if err != nil {
				panic(err)
			}
			
			frame, _ := OSFParseFrame(buf)
			frames = append(frames, *frame)

			fmt.Printf("Sample Number: %d\n", len(frames))

			if (len(frames) >= 300) {
				fmt.Printf("change!\n")
				break
			}
		}

		attributes = append(attributes, BuildAttributeList(class, frames)...)
		class += 1.0
	}

	for i := range attributes {
		j := rand.Intn(i + 1)
		attributes[i], attributes[j] = attributes[j], attributes[i]
	}
	
	model, _ := TrainExpressionsModel(attributes)

	for true {
		_, _, err := conn.ReadFrom(buf)
		if err != nil {
			panic(err)
		}
		
		frame, _ := OSFParseFrame(buf)
		slice := (*frame).TrainingSlice()
		x := make(map[int]float64, len(slice))
		for i, v := range slice {
			x[i+1] = v
		}

		attributes := make(map[int]string, 3)
		attributes[1] = "Neutral"
		attributes[2] = "Smiling"
		attributes[3] = "Pog"
		
		fmt.Printf("Label: %s\n", attributes[int(model.Predict(x))])
	}


	libSvm.NewParameter()
}