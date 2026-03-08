package slice

// ToSetBy returns a map with each element's key (extracted by fn) set to true.
// Useful when elements aren't directly comparable but have a comparable key field.
func ToSetBy[T any, K comparable](ts Mapper[T], fn func(T) K) map[K]bool {
	set := make(map[K]bool, len(ts))
	for _, t := range ts {
		set[fn(t)] = true
	}
	return set
}
