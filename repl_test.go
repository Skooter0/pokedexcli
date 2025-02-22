package main

import "testing"

func TestCleanInput(t *testing.T) {
    cases := []struct {
        input    string
        expected []string
    }{
        {
            input:    "  hello  world  ",
            expected: []string{"hello", "world"},
        },
        // Let's add another test case here
        {
            input:    "Charmander Bulbasaur PIKACHU",
            expected: []string{"charmander", "bulbasaur", "pikachu"},
        },
    }
    for _, c := range cases {
        actual := cleanInput(c.input)
        if len(actual) != len(c.expected) {
            t.Errorf("got len %d, want len %d", len(actual), len(c.expected))
            continue
        }
        for i := range actual {
            if actual[i] != c.expected[i] {
                t.Errorf("got %q, want %q", actual[i], c.expected[i])
            }
        }
    }
}