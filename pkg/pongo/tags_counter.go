package pongo

import (
	"fmt"
	"sync"

	"github.com/flosch/pongo2/v6"
)

// counterStorage stores the counters for counter tags
// Uses a map with mutex for thread-safety across template executions
var (
	counterStorage = make(map[string]int)
	counterMutex   sync.RWMutex
)

// counterNode represents the counter tag that increments a named counter
type counterNode struct {
	name string
}

// counterShowNode represents the counter_show tag that displays counter value(s)
type counterShowNode struct {
	names []string
}

// ResetCounters resets all counters (call before each template render)
func ResetCounters() {
	counterMutex.Lock()
	defer counterMutex.Unlock()

	counterStorage = make(map[string]int)
}

// GetCounter returns the current value of a named counter
func GetCounter(name string) int {
	counterMutex.RLock()
	defer counterMutex.RUnlock()

	return counterStorage[name]
}

// makeCounterTag creates the counter tag parser with expression support
// Usage: {% counter "1100" %}
func makeCounterTag() pongo2.TagParser {
	return func(_ *pongo2.Parser, _ *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
		if arguments.Remaining() == 0 {
			return nil, arguments.Error("counter tag requires a counter name", nil)
		}

		// Get the counter name (should be a string)
		token := arguments.Current()
		if token.Typ != pongo2.TokenString {
			return nil, arguments.Error("counter tag requires a string argument", nil)
		}

		name := token.Val

		arguments.Consume()

		return &counterNode{name: name}, nil
	}
}

// Execute increments the named counter
func (node *counterNode) Execute(_ *pongo2.ExecutionContext, _ pongo2.TemplateWriter) *pongo2.Error {
	counterMutex.Lock()
	defer counterMutex.Unlock()

	counterStorage[node.name]++

	return nil
}

// makeCounterShowTag creates the counter_show tag parser
// Usage: {% counter_show "1100" %} or {% counter_show "1100" "1101" "1102" %}
func makeCounterShowTag() pongo2.TagParser {
	return func(_ *pongo2.Parser, _ *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
		if arguments.Remaining() == 0 {
			return nil, arguments.Error("counter_show tag requires at least one counter name", nil)
		}

		var names []string

		// Parse all string arguments
		for arguments.Remaining() > 0 {
			token := arguments.Current()
			if token.Typ != pongo2.TokenString {
				break
			}

			names = append(names, token.Val)

			arguments.Consume()
		}

		if len(names) == 0 {
			return nil, arguments.Error("counter_show tag requires string arguments", nil)
		}

		return &counterShowNode{names: names}, nil
	}
}

// Execute displays the sum of all named counters
func (node *counterShowNode) Execute(_ *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
	counterMutex.RLock()
	defer counterMutex.RUnlock()

	total := 0
	for _, name := range node.names {
		total += counterStorage[name]
	}

	if _, err := fmt.Fprintf(writer, "%d", total); err != nil {
		return &pongo2.Error{
			Sender:    "counter_show",
			OrigError: err,
		}
	}

	return nil
}
