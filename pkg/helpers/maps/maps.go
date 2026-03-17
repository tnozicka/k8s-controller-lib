package maps

import gomaps "maps"

func HasKey[K, V comparable](m map[K]V, key K) bool {
	_, found := m[key]
	return found
}

// Merge merges several maps into a new map. If there are conflicting keys, the last one wins and overwrites the value.
func Merge[Key comparable, Value any](maps ...map[Key]Value) map[Key]Value {
	res := map[Key]Value{}
	for _, m := range maps {
		gomaps.Copy(res, m)
	}
	return res
}

func LookupValues[K comparable, T any](m map[K]T, keys ...K) []T {
	res := make([]T, 0, len(keys))

	for _, k := range keys {
		item, found := m[k]
		if !found {
			continue
		}

		res = append(res, item)
	}

	return res
}
