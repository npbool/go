package scoring

import (
	"sort"
	"fmt"
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
	fmt.Println(truth.rank_truth, pred.rank_prediction)
	return 0.666
}