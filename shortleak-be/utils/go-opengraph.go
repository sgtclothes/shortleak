package utils

import (
	"net/http"

	"github.com/dyatlov/go-opengraph/opengraph"
)

func GetOpenGraphData(url string) (opengraph.OpenGraph, error) {
	resp, err := http.Get(url)
	if err != nil {
		return opengraph.OpenGraph{}, err
	}
	defer resp.Body.Close()

	og := opengraph.NewOpenGraph()
	if err := og.ProcessHTML(resp.Body); err != nil {
		return opengraph.OpenGraph{}, err
	}
	return *og, nil
}
