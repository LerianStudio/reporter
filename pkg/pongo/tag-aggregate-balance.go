// Copyright (c) 2026 Lerian Studio. All rights reserved.
// Use of this source code is governed by the Elastic License 2.0
// that can be found in the LICENSE file.

package pongo

import (
	"fmt"
	"sort"
	"time"

	"github.com/LerianStudio/reporter/pkg/constant"

	"github.com/flosch/pongo2/v6"
	"github.com/shopspring/decimal"
)

// aggregateBalanceNode represents the aggregate_balance tag.
// It groups items by a primary field, sub-groups by account/route,
// selects the last item by date from each sub-group, and sums the balances.
type aggregateBalanceNode struct {
	collectionExpr   pongo2.IEvaluator // Expression for the collection to aggregate
	balanceFieldExpr pongo2.IEvaluator // Field containing the balance value
	groupByExpr      pongo2.IEvaluator // Field to group by (e.g., cosif_code)
	orderByExpr      pongo2.IEvaluator // Field to order by for selecting last item (e.g., created_at)
	filterExpr       pongo2.IEvaluator // Optional filter condition
	resultVarName    string            // Variable name to store the result
}

// AggregateBalanceItem represents a single aggregated result.
type AggregateBalanceItem struct {
	GroupValue string          `json:"group_value"` // Value of the group_by field
	Balance    decimal.Decimal `json:"balance"`     // Sum of last balances from each account
	Count      int             `json:"count"`       // Number of accounts/routes in this group
}

// makeAggregateBalanceTag creates the tag parser for aggregate_balance.
// Syntax: {% aggregate_balance <collection> by "<balance_field>" group_by "<group_field>" order_by "<date_field>" [if <condition>] as <result_var> %}
func makeAggregateBalanceTag() pongo2.TagParser {
	return func(doc *pongo2.Parser, start *pongo2.Token, args *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
		// 1. Parse collection expression
		collectionExpr, err := args.ParseExpression()
		if err != nil {
			return nil, err
		}

		// 2. Expect 'by' keyword + balance field
		if args.Match(pongo2.TokenIdentifier, "by") == nil {
			return nil, args.Error("Expected 'by' keyword", nil)
		}

		balanceFieldExpr, err := args.ParseExpression()
		if err != nil {
			return nil, err
		}

		// 3. Expect 'group_by' keyword + group field
		if args.Match(pongo2.TokenIdentifier, "group_by") == nil {
			return nil, args.Error("Expected 'group_by' keyword", nil)
		}

		groupByExpr, err := args.ParseExpression()
		if err != nil {
			return nil, err
		}

		// 4. Expect 'order_by' keyword + date field
		if args.Match(pongo2.TokenIdentifier, "order_by") == nil {
			return nil, args.Error("Expected 'order_by' keyword", nil)
		}

		orderByExpr, err := args.ParseExpression()
		if err != nil {
			return nil, err
		}

		// 5. Optional 'if' condition
		var filterExpr pongo2.IEvaluator
		if args.Match(pongo2.TokenIdentifier, "if") != nil {
			filterExpr, err = args.ParseExpression()
			if err != nil {
				return nil, err
			}
		}

		// 6. Expect 'as' keyword + result variable name
		// Note: 'as' is a reserved keyword in pongo2, so we use TokenKeyword
		if args.Match(pongo2.TokenKeyword, "as") == nil {
			return nil, args.Error("Expected 'as' keyword", nil)
		}

		resultVarToken := args.MatchType(pongo2.TokenIdentifier)
		if resultVarToken == nil {
			return nil, args.Error("Expected variable name after 'as'", nil)
		}

		return &aggregateBalanceNode{
			collectionExpr:   collectionExpr,
			balanceFieldExpr: balanceFieldExpr,
			groupByExpr:      groupByExpr,
			orderByExpr:      orderByExpr,
			filterExpr:       filterExpr,
			resultVarName:    resultVarToken.Val,
		}, nil
	}
}

// Execute processes the aggregate_balance tag.
func (node *aggregateBalanceNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
	// 1. Evaluate collection
	list, err := evaluateCollection(ctx, node.collectionExpr)
	if err != nil {
		return err
	}

	// 2. Check collection size to prevent resource exhaustion
	if len(list) > constant.MaxAggregateBalanceCollectionSize {
		return ctx.Error(fmt.Sprintf("collection size %d exceeds maximum allowed %d", len(list), constant.MaxAggregateBalanceCollectionSize), nil)
	}

	// 2. Get field names
	balanceField := node.getFieldName(ctx, node.balanceFieldExpr)
	groupByField := node.getFieldName(ctx, node.groupByExpr)
	orderByField := node.getFieldName(ctx, node.orderByExpr)

	// 3. Filter items (if condition present)
	filtered := node.filterItems(ctx, list)

	// 4. Group by primary field (e.g., cosif_code)
	groups := node.groupByField(filtered, groupByField)

	// 5. Process each group
	results := node.processGroups(groups, balanceField, orderByField)

	// 6. Set result variable in context
	ctx.Private[node.resultVarName] = results

	return nil
}

// getFieldName evaluates a field expression and returns its string value.
func (node *aggregateBalanceNode) getFieldName(ctx *pongo2.ExecutionContext, expr pongo2.IEvaluator) string {
	if expr == nil {
		return ""
	}

	val, err := expr.Evaluate(ctx)
	if err != nil {
		return ""
	}

	return val.String()
}

// filterItems applies the filter expression to each item and returns matching items.
func (node *aggregateBalanceNode) filterItems(ctx *pongo2.ExecutionContext, items []map[string]any) []map[string]any {
	if node.filterExpr == nil {
		return items
	}

	var filtered []map[string]any

	for _, item := range items {
		if passesFilter(ctx, item, node.filterExpr) {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

// groupByField groups items by the specified field.
func (node *aggregateBalanceNode) groupByField(items []map[string]any, field string) map[string][]map[string]any {
	groups := make(map[string][]map[string]any)

	for _, item := range items {
		if val, ok := getNestedField(item, field); ok {
			key := fmt.Sprintf("%v", val)
			groups[key] = append(groups[key], item)
		}
	}

	return groups
}

// processGroups processes each group to get aggregated balances.
func (node *aggregateBalanceNode) processGroups(groups map[string][]map[string]any, balanceField, orderByField string) []map[string]any {
	results := make([]map[string]any, 0, len(groups))

	for groupValue, items := range groups {
		// Sub-group by account/route
		subGroups := node.subGroupByAccount(items)

		// Get last balance from each sub-group and sum
		totalBalance := decimal.Zero
		count := 0

		for _, subItems := range subGroups {
			lastItem := node.getLastByDate(subItems, orderByField)
			if lastItem != nil {
				if balance := node.extractBalance(lastItem, balanceField); balance != nil {
					totalBalance = totalBalance.Add(*balance)
					count++
				}
			}
		}

		results = append(results, map[string]any{
			"group_value": groupValue,
			"balance":     totalBalance,
			"count":       count,
		})
	}

	// Sort results by group_value for deterministic output
	sort.Slice(results, func(i, j int) bool {
		vi, okI := results[i]["group_value"].(string)
		vj, okJ := results[j]["group_value"].(string)

		if !okI || !okJ {
			return false // stable ordering for invalid entries
		}

		return vi < vj
	})

	return results
}

// subGroupByAccount sub-groups items by account_id or route_id.
func (node *aggregateBalanceNode) subGroupByAccount(items []map[string]any) map[string][]map[string]any {
	subGroups := make(map[string][]map[string]any)

	for _, item := range items {
		key := node.getSubGroupKey(item)
		subGroups[key] = append(subGroups[key], item)
	}

	return subGroups
}

// getSubGroupKey determines the sub-group key for an item.
// Priority: account_id > route_id > id > "_default_"
func (node *aggregateBalanceNode) getSubGroupKey(item map[string]any) string {
	if val, ok := getNestedField(item, "account_id"); ok {
		return fmt.Sprintf("%v", val)
	}

	if val, ok := getNestedField(item, "route_id"); ok {
		return fmt.Sprintf("%v", val)
	}

	if val, ok := getNestedField(item, "id"); ok {
		return fmt.Sprintf("%v", val)
	}

	return "_default_"
}

// getLastByDate returns the item with the latest date.
// If multiple items have the same date, uses the last one in iteration order.
// Items without a date field are included with zero time (lowest priority).
func (node *aggregateBalanceNode) getLastByDate(items []map[string]any, dateField string) map[string]any {
	if len(items) == 0 {
		return nil
	}

	var latest map[string]any

	var latestTime time.Time

	for _, item := range items {
		var itemTime time.Time

		if val, ok := getNestedField(item, dateField); ok {
			itemTime = node.parseTime(val)
		}
		// Items without date field get zero time (will be overwritten by any dated item)

		// Use >= to ensure deterministic behavior: later items win ties
		if latest == nil || itemTime.After(latestTime) || itemTime.Equal(latestTime) {
			latest = item
			latestTime = itemTime
		}
	}

	return latest
}

// parseTime attempts to parse a value as time.Time.
func (node *aggregateBalanceNode) parseTime(val any) time.Time {
	switch v := val.(type) {
	case time.Time:
		return v
	case string:
		// Try RFC3339 first
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			return t
		}
		// Try date only
		if t, err := time.Parse("2006-01-02", v); err == nil {
			return t
		}
		// Try RFC3339Nano
		if t, err := time.Parse(time.RFC3339Nano, v); err == nil {
			return t
		}
	}

	return time.Time{}
}

// extractBalance extracts a decimal balance from an item's field.
func (node *aggregateBalanceNode) extractBalance(item map[string]any, field string) *decimal.Decimal {
	val, ok := getNestedField(item, field)
	if !ok {
		return nil
	}

	var d decimal.Decimal

	switch v := val.(type) {
	case int:
		d = decimal.NewFromInt(int64(v))
	case int64:
		d = decimal.NewFromInt(v)
	case float64:
		d = decimal.NewFromFloat(v)
	case string:
		parsed, err := decimal.NewFromString(v)
		if err != nil {
			return nil
		}

		d = parsed
	case decimal.Decimal:
		d = v
	default:
		return nil
	}

	return &d
}
