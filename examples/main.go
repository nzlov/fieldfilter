package main

import (
	"encoding/json"
	"fmt"

	"github.com/nzlov/fieldfilter"
)

type A struct {
	Name  string      `json:"name"`
	Sex   int64       `json:"sex"`
	Inter interface{} `json:"inter"`
	A     *A          `json:"a"`
	As    []A         `json:"as"`
	Is    []int       `json:"is"`
	Isp   []*int      `json:"isp"`

	Lanaguages fieldfilter.Lanaguages
}

type As struct {
	A A `filter:"*"`
}

func (a A) Lang() fieldfilter.Lanaguages {
	return a.Lanaguages
}

func main() {
	v := 1

	a := A{
		Name:  "a",
		Sex:   6,
		Inter: 66,
		Is:    []int{1, 2},
		Isp:   []*int{&v},
		A: &A{
			Name: "b",
			A: &A{
				Name: "bb",
			},
			Is: []int{},
			Inter: map[string]int{
				"a": 1,
			},
		},
		As: []A{
			{
				Name: "bs0",
				Lanaguages: fieldfilter.Lanaguages{
					"en": {
						"name": "bs0_en",
					},
				},
			},
			{
				Name: "bs1",
			},
		},
		Lanaguages: fieldfilter.Lanaguages{
			"en": {
				"name":  "a_en",
				"sex":   "v",
				"inter": "xx",
			},
			"es": {
				"name": "a_es",
			},
		},
	}

	fm := fieldfilter.FiltersToMap("name,inter,is,a.name,a.inter")

	data, _ := json.MarshalIndent(fieldfilter.FilterLang("en", a, "", fm), "", "  ")

	fmt.Println(string(data))
}

