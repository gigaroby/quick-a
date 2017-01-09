package model

type Category struct {
	Index int    `json:"index"`
	Name  string `json:"name"`
}

type Categories []Category
