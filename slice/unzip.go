package slice

// Unzip2 extracts two slices from ts in a single pass by applying the extraction functions.
// This is more efficient than calling two separate mapping operations when you need multiple fields.
func Unzip2[T, A, B any](ts []T, fa func(T) A, fb func(T) B) (Mapper[A], Mapper[B]) {
	as := make([]A, len(ts))
	bs := make([]B, len(ts))
	for i, t := range ts {
		as[i] = fa(t)
		bs[i] = fb(t)
	}

	return as, bs
}

// Unzip3 extracts three slices from ts in a single pass by applying the extraction functions.
func Unzip3[T, A, B, C any](ts []T, fa func(T) A, fb func(T) B, fc func(T) C) (Mapper[A], Mapper[B], Mapper[C]) {
	as := make([]A, len(ts))
	bs := make([]B, len(ts))
	cs := make([]C, len(ts))
	for i, t := range ts {
		as[i] = fa(t)
		bs[i] = fb(t)
		cs[i] = fc(t)
	}

	return as, bs, cs
}

// Unzip4 extracts four slices from ts in a single pass by applying the extraction functions.
func Unzip4[T, A, B, C, D any](ts []T, fa func(T) A, fb func(T) B, fc func(T) C, fd func(T) D) (Mapper[A], Mapper[B], Mapper[C], Mapper[D]) {
	as := make([]A, len(ts))
	bs := make([]B, len(ts))
	cs := make([]C, len(ts))
	ds := make([]D, len(ts))
	for i, t := range ts {
		as[i] = fa(t)
		bs[i] = fb(t)
		cs[i] = fc(t)
		ds[i] = fd(t)
	}

	return as, bs, cs, ds
}
