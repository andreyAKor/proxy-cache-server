package get

import (
	"testing"
)

func TestRoundRobin(t *testing.T) {
	type testpair struct {
		params map[string]
		result string
	}
//"http://163.172.215.220:80", "http://82.7.253.143:80", "https://89.236.17.108:3128"
	tests := []testpair{
		{"", ""},
		{
			map[string]{"http://163.172.215.220:80", "http://82.7.253.143:80", "https://89.236.17.108:3128"},
			"http://163.172.215.220:80"
		}
	}

	for _, pair := range tests {
		v := RoundRobin(pair.value)

		if v != pair.result {
			t.Error(
				"For", pair.value,
				"expected", pair.result,
				"got", v,
			)
		}
	}
}
