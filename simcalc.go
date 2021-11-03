package main

import "strings"

// https://www.golangprograms.com/golang-program-for-implementation-of-levenshtein-distance.html
func levenshtein(str1, str2 []rune) int {
	s1len := len(str1)
	s2len := len(str2)
	column := make([]int, len(str1)+1)

	for y := 1; y <= s1len; y++ {
		column[y] = y
	}
	for x := 1; x <= s2len; x++ {
		column[0] = x
		lastkey := x - 1
		for y := 1; y <= s1len; y++ {
			oldkey := column[y]
			var incr int
			if str1[y-1] != str2[x-1] {
				incr = 1
			}

			column[y] = minimum(column[y]+1, column[y-1]+1, lastkey+incr)
			lastkey = oldkey
		}
	}
	return column[s1len]
}

func minimum(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
	} else {
		if b < c {
			return b
		}
	}
	return c
}

func similarity(s1, s2 string) float64 {
	sumlen := len(s1) + len(s2)
	if sumlen == 0 {
		return 0
	}
	l := levenshtein([]rune(strings.ToUpper(s1)), []rune(strings.ToUpper(s2)))
	return float64(sumlen-l) / float64(sumlen)
}

func runeEq(arr1, arr2 []rune) bool {
	if len(arr1) != len(arr2) {
		return false
	}
	for i, _ := range arr1 {
		if arr1[i] != arr2[i] {
			return false
		}
	}
	return true
}

type RL []rune
type RLArray []RL
type RLMap map[int]RLArray

func rl2map(rr RL) RLMap {
	res := make(RLMap)
	ll := len(rr)
	if ll > 3 {
		ll = 3
	}
	for l := 1; l <= ll; l++ {
		for i := 0; i < len(rr)-l+1; i++ {
			res[l] = append(res[l], rr[i:i+l])
		}
	}
	return res
}

//return percent [0..1] rm1 in rm2
func mapSimilarity(rm1, rm2 RLMap) float64 {
	var up float64 = 0
	var down float64 = 0
	for l, lst1 := range rm1 {
		down += float64(1 * len(lst1))
		lst2, ok := rm2[l]
		if !ok {
			continue
		}
		for _, a1 := range lst1 {
			for _, a2 := range lst2 {
				if runeEq(a1, a2) {
					up += float64(1)
					break
				}
			}
		}
	}
	if down == 0 {
		return 0
	}
	return up / down
}

func similarity2(s1, s2 string) float64 {
	r1 := rl2map([]rune(strings.ToUpper(s1)))
	r2 := rl2map([]rune(strings.ToUpper(s2)))
	return mapSimilarity(r1, r2)
}
