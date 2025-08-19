package voice

import "math"

func rmsEnergy(f []float32) float64 {
	var s float64
	for _, v := range f {
		s += float64(v * v)
	}
	if len(f) == 0 {
		return 0
	}
	return math.Sqrt(s / float64(len(f)))
}

func i16ToF32(in []int16, out []float32) {
	for i := range in {
		out[i] = float32(in[i]) / 32768.0
	}
}

func flatten(frames [][]float32) []float32 {
	var total int
	for _, f := range frames {
		total += len(f)
	}
	out := make([]float32, 0, total)
	for _, f := range frames {
		out = append(out, f...)
	}
	return out
}
