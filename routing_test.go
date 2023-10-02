package rabbitmq

import "testing"

type routingKeyPairs struct {
	n string // needle
	s string // stack(received routing key)
}

func TestPositive(t *testing.T) {
	var cases = []routingKeyPairs{
		{"change.user", "change.user"},
		{"*.user", "change.user"},
		{"*.user.*", "change.user.uss"},
		{"change.#", "change.user"},
		{"change.#", "change.user.uss"},
		{"#.user.#", "change.user.uss"},
		{"change.#.uss", "change.user.420.uss"},
		{"*.user.#", "change.user.uss"},
		{"#.user.*", "change.user.uss"},
		{"change.#.uss.*", "change.user.420.uss.host1"},
	}
	for _, c := range cases {
		if !MatchKey(c.n, c.s) {
			t.Errorf("Expect match needle(%s) to routing key(%s)", c.n, c.s)
		}
	}
}

func TestNegative(t *testing.T) {
	var cases = []routingKeyPairs{
		{"change.user", "delete.user"},
		{"change.user", "change.user.uss"},
		{"change.user.*", "change.user"},
		// TODO: move to positive!
		{"change.user.#", "change.user"},
	}
	for _, c := range cases {
		if MatchKey(c.n, c.s) {
			t.Errorf("Expect not match needle(%s) to routing key(%s)", c.n, c.s)
		}
	}
}
