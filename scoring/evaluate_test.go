package scoring

import (
	"testing"
)

func Test_AUC_Recall_Precision_F(t *testing.T) {
	truth_cla := map[string]int {
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

	prediction_cla := map[string]float32 {
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

	truth := Truth{}
	prediction := Prediction{}
	truth.classification_truth = truth_cla
	prediction.classification_prediction = prediction_cla

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

func Test_MAP_NDCG(t *testing.T) {
	truth := Truth{}
	prediction := Prediction{}
	
	truth.rank_truth = []rank {
		{1,2,3,4,5},
		{1,2,3,4,5},
	}
	prediction.rank_prediction = []rank {
		{6,4,7,1,2},
		{1,1,1,1,1},
	}

	if MAP(truth, prediction) == 0.26 {
		t.Log("MAP Correct")
	} else {
		t.Error("MAP Sucked")
	}

	truth.rank_truth = []rank {
		{5,4,3,2,1,0},
		{5,4,3,2,1,0},
	}
	prediction.rank_prediction = []rank {
		{4,1,2,0,3,5},
		{1,0,2,3,4,5},
	}
	if NDCG(truth, prediction) == 0.53729945 {
		t.Log("NDCG Correct")
	} else {
		t.Error("NDCG Sucked")
	}
}