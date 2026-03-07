package config

import "testing"

func TestGetEnvInt(t *testing.T) {
	t.Run("value exists", func(t *testing.T) {
		t.Setenv("TEST_INT", "12")
		if got := GetEnvInt("TEST_INT", 1); got != 12 {
			t.Fatalf("got = %d", got)
		}
	})

	t.Run("value missing", func(t *testing.T) {
		t.Setenv("TEST_INT", "")
		if got := GetEnvInt("TEST_INT", 5); got != 5 {
			t.Fatalf("got = %d", got)
		}
	})

	t.Run("value invalid", func(t *testing.T) {
		t.Setenv("TEST_INT", "bad")
		if got := GetEnvInt("TEST_INT", 9); got != 9 {
			t.Fatalf("got = %d", got)
		}
	})
}
