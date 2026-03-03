package slices

func Convert[To, From any](convert func(From) To, objs ...From) []To {
	res := make([]To, 0, len(objs))

	for i := range objs {
		res = append(res, convert(objs[i]))
	}

	return res
}
