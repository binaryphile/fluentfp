package slice

type Any = Mapper[any]
type Bool = Mapper[bool]
type Byte = Mapper[byte]
type Error = Mapper[error]
type Float32 = Mapper[float32]
type Float64 []float64

// Sum returns the sum of all elements.
func (fs Float64) Sum() float64 {
	var sum float64
	for _, f := range fs {
		sum += f
	}
	return sum
}
type Int = Mapper[int]
type Rune = Mapper[rune]
type String = Mapper[string]
