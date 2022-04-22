package main

import (
	"log"
	"strings"
	"testing"
)

func DoNothing(a, b string) {

}

func Benchmark_StringSeparate(b *testing.B) {
	b.ReportAllocs()

	str := "abc=123:qwe=omg:ppp=connected"

	for i := 0; i < b.N; i++ {
		a := strings.Split(str, ":")

		for _, v := range a {
			z := strings.Split(v, "=")

			DoNothing(z[0], z[1])
		}
	}
}

func Benchmark_StringParser(b *testing.B) {
	b.ReportAllocs()

	str := "abc=123:qwe=omg:ppp=connected"

	for i := 0; i < b.N; i++ {
		var prev int

		for z := 0; z <= len(str); z++ {
			if z == len(str) || str[z] == ':' {
				substr := str[prev:z]

				for e := 0; e < len(substr); e++ {
					if substr[e] == '=' {
						aS := substr[:e]
						bS := substr[e+1:]

						DoNothing(aS, bS)
					}
				}

				z++
				prev = z
			}
		}
	}
}

func Test_StringParser(t *testing.T) {
	str := "abc=123:qwe=omg:ppp=connected"

	var prev int

	for z := 0; z <= len(str); z++ {
		if z == len(str) || str[z] == ':' {
			substr := str[prev:z]

			for e := 0; e < len(substr); e++ {
				if substr[e] == '=' {
					a := substr[:e]
					b := substr[e+1:]

					log.Println(a, b)
				}
			}

			z++
			prev = z
		}
	}
}
