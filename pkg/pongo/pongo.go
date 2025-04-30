package pongo

import (
	"fmt"
	"github.com/flosch/pongo2/v6"
)

// init initializes custom filters and tags for the Pongo2 template engine. It registers filters and aggregation tags.
func init() {
	if err := pongo2.RegisterFilter("scale", scaleFilter); err != nil {
		panic("Failed to register scale filter: " + err.Error())
	}

	if err := pongo2.RegisterFilter("percent_of", percentOfFilter); err != nil {
		panic("Failed to register percent_of filter: " + err.Error())
	}

	if err := pongo2.RegisterFilter("slice_str", sliceFilter); err != nil {
		panic("Failed to register slice filter: " + err.Error())
	}

	tags := []struct {
		name string
		op   string
	}{
		{"sum_by", "sum"},
		{"count_by", "count"},
		{"avg_by", "avg"},
		{"min_by", "min"},
		{"max_by", "max"},
	}

	for _, tag := range tags {
		if err := pongo2.RegisterTag(tag.name, makeAggregateTag(tag.op)); err != nil {
			panic(fmt.Sprintf("Failed to register tag '%s': %s", tag.name, err.Error()))
		}
	}
}
