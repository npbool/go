package scoring

import (
	"sort"
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

func AUC(truth Truth, pred map[string]float32) float32 {
	ins := make(IList, len(truth))
	i := 0
	numPos := 0
	numNeg := 0
	for k, _ := range truth {
		ins[i] = Instance{truth[k], pred[k]}
		i++
		if truth[k] == 0 {
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
