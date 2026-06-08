package score

import "testing"

func TestComputePassAtK(t *testing.T) {
	rounds := []RoundResult{
		{Round: 1, Verifier: "fail"},
		{Round: 2, Verifier: "pass"},
		{Round: 3, Verifier: "fail"},
	}
	if got := ComputePassAtK(rounds, 1); got != 0 {
		t.Fatalf("pass@1 = %v", got)
	}
	if got := ComputePassAtK(rounds, 3); got != 1 {
		t.Fatalf("pass@3 = %v", got)
	}
}
