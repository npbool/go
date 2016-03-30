package scoring

import (
	"sort"
	"fmt"
	"math"
)

type Instance struct {
	label int
	score float32
}

type IList []Instance

func (ins IList) Len() int {
	return len(ins)
}
func (ins IList) Swap(i, j int) {
	ins[i], ins[j] = ins[j], ins[i]
}
func (ins IList) Less(i, j int) bool {
	return ins[i].score >= ins[j].score
}

func AUC(truth Truth, pred Prediction) float32 {
	ins := make(IList, len(truth.classification_truth))
	i := 0
	numPos := 0
	numNeg := 0
	for k, _ := range truth.classification_truth {
		ins[i] = Instance{truth.classification_truth[k], pred.classification_prediction[k]}
		i++
		if truth.classification_truth[k] == 0 {
			numNeg++
		} else {
			numPos++
		}
	}
	sort.Sort(ins)

	auc := 0.0
	tpr := 0.0
	fpr := 0.0
	tp := 0
	fp := 0
	for i := 0; i < len(ins); {
		if i > 0 && ins[i].score != ins[i-1].score {
			newTpr := float64(tp) / float64(numPos)
			newFpr := float64(fp) / float64(numNeg)

			auc += (newFpr - fpr) * (newTpr + tpr) / 2
			fpr = newFpr
			tpr = newTpr
		}

		if ins[i].label == 0 {
			fp++
		} else {
			tp++
		}

		i++
	}
	auc += (1 - fpr) * (1 + tpr) / 2
	return float32(auc)
}

func calculatePara(truth Truth, pred Prediction) (int, int, int, int) {
	ins := make(IList, len(truth.classification_truth))
	i := 0
	for k, _ := range truth.classification_truth {
		ins[i] = Instance{truth.classification_truth[k], pred.classification_prediction[k]}
		i++
	}

	tp := 0
	tn := 0
	fp := 0
	fn := 0
	for i := 0; i < len(ins); i++ {
		if ins[i].label == 1 {
			if ins[i].score == 1 {
				tp++
			}
			if ins[i].score == 0 {
				fn++
			}
		} else {
			if ins[i].score == 1 {
				fp++
			}
			if ins[i].score == 0 {
				tn++
			}
		}
	}

	return tp, tn, fp, fn
}

func divisorCheck(divisor float32) (error) {
	if divisor == 0 {
		return fmt.Errorf("%f can't divided by 0", divisor)
	}
	return nil
}

func Recall(truth Truth, pred Prediction) float32 {
	tp, _, _, fn := calculatePara(truth, pred)

	if err := divisorCheck(float32(tp + fn)); err != nil {
		fmt.Println(err)
		return 0
	}
	return float32(tp) / float32(tp + fn)
}

func Precision(truth Truth, pred Prediction) float32 {
	tp, _, fp, _ := calculatePara(truth, pred)

	if err := divisorCheck(float32(tp + fp)); err != nil {
		fmt.Println(err)
		return 0
	}
	return float32(tp) / float32(tp + fp)
}

func F_score(truth Truth, pred Prediction) float32 {
	recall := Recall(truth, pred)
	precision := Precision(truth, pred)

	if err := divisorCheck(float32(precision + recall)); err != nil {
		fmt.Println(err)
		return 0
	}
	return float32(2 * precision * recall) / float32(precision + recall)
}

func MAP(truth Truth, pred Prediction) float32 {
	if len(truth.rank_truth) == 0 {
		return 0
	}
	sum := float32(0)
	num_line := len(truth.rank_truth)

	for i := 0; i < num_line; i++ {
		if len(truth.rank_truth[i]) == 0 {
			continue
		}

		score := float32(0)
		num_hits := float32(0)

		for pred_index, pred_value := range pred.rank_prediction[i] {
			flag1 := 0
			for _, truth_value := range truth.rank_truth[i] {
				if truth_value == pred_value {
					flag1 = 1
					break
				}
			}
			flag2 := 0
			for _, pre_pred_value := range pred.rank_prediction[i][:pred_index] {
				if pre_pred_value == pred_value {
					flag2 = 1
					break
				}
			}

			if flag1 == 1 && flag2 == 0 {
				num_hits += 1
				score += float32(num_hits) / float32(pred_index+1)
			}
		}
		sum += float32(score) / float32(len(truth.rank_truth[i]))
	}
	return sum / float32(len(truth.rank_truth))
}

func DCG(truth_rank, pred_rank rank) float32 {
	k := len(pred_rank)
	pre_pred_rank := make([]int, k)
	
	for i := 0; i < k; {
		pre_pred_rank[i] = pred_rank[k - i - 1]
		i++
	}
	// fmt.Println(pred_rank, pre_pred_rank)

	sum := float64(0)
	rels := make([]int, k)
	gains := make([]float64, k)
	discounts := make([]float64, k)
	for i := 0; i < k; {
		rels[i] = truth_rank[pre_pred_rank[i]]
		gains[i] = math.Pow(float64(2), float64(rels[i])) - float64(1)
		discounts[i] = math.Log2(float64(i + 2))
		sum += gains[i] / discounts[i]
		i++
	}
	// fmt.Println(rels, gains, discounts, sum)
	return float32(sum)
}

func NDCG(truth Truth, pred Prediction) float32 {
	if len(truth.rank_truth) == 0 {
		return 0
	}
	sum := float32(0)
	num_line := len(truth.rank_truth)

	for i := 0; i < num_line; i++ {
		// fmt.Println(truth.rank_truth[i], pred.rank_prediction[i])
		actual_score := DCG(truth.rank_truth[i], pred.rank_prediction[i])
		best_score := DCG(truth.rank_truth[i], truth.rank_truth[i])
		sum += float32(actual_score) / float32(best_score)
	}
	return sum / float32(len(truth.rank_truth))
}