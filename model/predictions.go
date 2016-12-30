package model

import (
	"sort"
)

type Prediction struct {
	CategoryIndex int     `json:"category_index"`
	CategoryName  string  `json:"category_name"`
	Confidence    float64 `json:"confidence"`
}

type Predictions []Prediction

func (p Predictions) Len() int {
	return len(p)
}

func (p Predictions) Less(i, j int) bool {
	return p[i].Confidence < p[j].Confidence
}

func (p Predictions) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p Predictions) Top(k int) Predictions {
	sort.Sort(sort.Reverse(p))
	return Predictions(p[:k])
}

func (p Predictions) FillNames(metadata map[int]string) Predictions {
	for i, _ := range p {
		p[i].CategoryName = metadata[p[i].CategoryIndex]
	}
	return p
}
