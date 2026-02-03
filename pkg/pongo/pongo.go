// Copyright (c) 2025 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package pongo

import (
	"fmt"

	"github.com/flosch/pongo2/v6"
)

// init initializes custom filters and tags for the Pongo2 template engine. It registers filters and aggregation tags.
func init() {
	if err := pongo2.RegisterFilter("percent_of", percentOfFilter); err != nil {
		panic("Failed to register percent_of filter: " + err.Error())
	}

	if err := pongo2.RegisterFilter("slice_str", sliceFilter); err != nil {
		panic("Failed to register slice filter: " + err.Error())
	}

	if err := pongo2.RegisterFilter("strip_zeros", stripZerosFilter); err != nil {
		panic("Failed to register strip_zeros filter: " + err.Error())
	}

	if err := pongo2.RegisterFilter("replace", replaceFilter); err != nil {
		panic("Failed to register replace filter: " + err.Error())
	}

	if err := pongo2.RegisterFilter("where", whereFilter); err != nil {
		panic("Failed to register where filter: " + err.Error())
	}

	if err := pongo2.RegisterFilter("sum", sumFilter); err != nil {
		panic("Failed to register sum filter: " + err.Error())
	}

	if err := pongo2.RegisterFilter("count", countFilter); err != nil {
		panic("Failed to register count filter: " + err.Error())
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
		{"date_time", "date"},
	}

	for _, tag := range tags {
		var err error

		if tag.op == "date" {
			err = pongo2.RegisterTag(tag.name, makeDateNowTag())
		} else {
			err = pongo2.RegisterTag(tag.name, makeAggregateTag(tag.op))
		}

		if err != nil {
			panic(fmt.Sprintf("Failed to register tag '%s': %s", tag.name, err.Error()))
		}
	}

	if err := pongo2.RegisterTag("calc", makeCalcTag); err != nil {
		panic("Failed to register calc tag: " + err.Error())
	}

	if err := pongo2.RegisterTag("aggregate_balance", makeAggregateBalanceTag()); err != nil {
		panic("Failed to register aggregate_balance tag: " + err.Error())
	}

	// Register counter tags for counting blocks during rendering
	if err := pongo2.RegisterTag("counter", makeCounterTag()); err != nil {
		panic("Failed to register counter tag: " + err.Error())
	}

	if err := pongo2.RegisterTag("counter_show", makeCounterShowTag()); err != nil {
		panic("Failed to register counter_show tag: " + err.Error())
	}
}
