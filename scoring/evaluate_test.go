package scoring

import (
	"testing"
)

func Test_AUC(t *testing.T) {
	truth := map[string]int {
		"8":  0,
		"17": 1,
		"27": 1,
		"7":  1,
		"10": 0,
		"16": 0,
		"18": 0,
		"29": 1,
		"35": 1,
		"9":  1,
		"5":  1,
		"11": 1,
		"15": 1,
		"28": 0,
		"2":  0,
		"14": 0,
		"22": 1,
		"25": 0,
		"31": 0,
		"6":  0,
	}

	prediction := map[string]float32 {
		"6":  1,
		"7":  1,
		"10": 0,
		"15": 1,
		"31": 0,
		"9":  0,
		"11": 1,
		"14": 1,
		"18": 0,
		"25": 0,
		"28": 1,
		"2":  0,
		"8":  0,
		"17": 1,
		"27": 1,
		"5":  1,
		"16": 0,
		"22": 0,
		"29": 1,
		"35": 1,
	}
	if AUC(truth, prediction) == 0.75 {
		t.Log("AUC Correct")
	} else {
		t.Error("AUC Sucked")
	}
	if Recall(truth, prediction) == 0.8 {
		t.Log("Recall Correct")
	} else {
		t.Error("Recall Sucked")
	}
	if Precision(truth, prediction) == 0.72727275 {
		t.Log("Precision Correct")
	} else {
		t.Error("Precision Sucked")
	}
	if F_score(truth, prediction) == 0.76190484 {
		t.Log("F_score Correct")
	} else {
		t.Error("F_score Sucked")
	}
}