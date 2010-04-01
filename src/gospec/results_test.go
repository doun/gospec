// Copyright © 2009-2010 Esko Luontola <www.orfjackal.net>
// This software is released under the Apache License 2.0.
// The license text is at http://www.apache.org/licenses/LICENSE-2.0

package gospec

import (
	"bytes"
	"testing"
)


func Test__When_results_have_many_root_specs__Then_they_are_sorted_alphabetically(t *testing.T) {
	results := newResultCollector()

	// register in reverse order
	a1 := newSpecRun("RootSpec2", nil, nil, nil)
	results.Update(a1)

	b2 := newSpecRun("RootSpec1", nil, nil, nil)
	results.Update(b2)

	// expect roots to be in alphabetical order
	assertReportIs(results, `
- RootSpec1
- RootSpec2

2 specs, 0 failures
`,
		t)
}

func Test__When_results_have_many_child_specs__Then_they_are_sorted_by_their_declaration_order(t *testing.T) {
	results := newResultCollector()

	// In tests, when a spec has many children, make sure
	// to pass a common parent instance to all the siblings.
	// Otherwise the parent's numberOfChildren is not
	// incremented and the children's paths will be wrong.

	// use names which would not sort alphabetically
	root := newSpecRun("RootSpec", nil, nil, nil)
	child1 := newSpecRun("one", nil, root, nil)
	child2 := newSpecRun("two", nil, root, nil)
	child3 := newSpecRun("three", nil, root, nil)

	// register in random order
	results.Update(root)
	results.Update(child1)

	results.Update(root)
	results.Update(child3)

	results.Update(root)
	results.Update(child2)

	// expect children to be in declaration order
	assertReportIs(results, `
- RootSpec
  - one
  - two
  - three

4 specs, 0 failures
`,
		t)
}

func Test__Collecting_results_of_zero_specs(t *testing.T) {
	results := newResultCollector()

	assertReportIs(results, `
0 specs, 0 failures
`,
		t)
}

func Test__Collecting_results_of_a_spec_with_no_children(t *testing.T) {
	results := newResultCollector()

	a1 := newSpecRun("RootSpec", nil, nil, nil)
	results.Update(a1)

	assertReportIs(results, `
- RootSpec

1 specs, 0 failures
`,
		t)
}

func Test__Collecting_results_of_a_spec_with_a_child(t *testing.T) {
	results := newResultCollector()

	a1 := newSpecRun("RootSpec", nil, nil, nil)
	a2 := newSpecRun("Child A", nil, a1, nil)
	results.Update(a1)
	results.Update(a2)

	assertReportIs(results, `
- RootSpec
  - Child A

2 specs, 0 failures
`,
		t)
}

func Test__Collecting_results_of_a_spec_with_nested_children(t *testing.T) {
	results := newResultCollector()

	a1 := newSpecRun("RootSpec", nil, nil, nil)
	a2 := newSpecRun("Child A", nil, a1, nil)
	a3 := newSpecRun("Child AA", nil, a2, nil)
	results.Update(a1)
	results.Update(a2)
	results.Update(a3)

	assertReportIs(results, `
- RootSpec
  - Child A
    - Child AA

3 specs, 0 failures
`,
		t)
}

func Test__Collecting_results_of_a_spec_with_multiple_nested_children(t *testing.T) {
	runner := NewRunner()
	runner.AddSpec("DummySpecWithMultipleNestedChildren", DummySpecWithMultipleNestedChildren)
	runner.Run()

	assertReportIs(runner.Results(), `
- DummySpecWithMultipleNestedChildren
  - Child A
    - Child AA
    - Child AB
  - Child B
    - Child BA
    - Child BB
    - Child BC

8 specs, 0 failures
`,
		t)
}

func Test__Collecting_results_of_failing_specs(t *testing.T) {
	results := newResultCollector()

	a1 := newSpecRun("Failing", nil, nil, nil)
	a1.AddError(newError("X did not equal Y", currentLocation()))
	results.Update(a1)

	b1 := newSpecRun("Passing", nil, nil, nil)
	b2 := newSpecRun("Child failing", nil, b1, nil)
	b2.AddError(newError("moon was not cheese", currentLocation()))
	results.Update(b1)
	results.Update(b2)

	assertReportIs(results, `
- Failing [FAIL]
    X did not equal Y
- Passing
  - Child failing [FAIL]
      moon was not cheese

3 specs, 2 failures
`,
		t)
}

func Test__When_spec_passes_on_first_run_but_fails_on_second__Then_the_error_is_reported(t *testing.T) {
	i := 0
	runner := NewRunner()
	runner.AddSpec("RootSpec", func(c Context) {
		if i == 1 {
			c.Then(10).Should.Equal(20)
		}
		i++
		c.Specify("Child A", func() {})
		c.Specify("Child B", func() {})
	})
	runner.Run()

	assertReportIs(runner.Results(), `
- RootSpec [FAIL]
    Expected '20' but was '10'
  - Child A
  - Child B

3 specs, 1 failures
`,
		t)
}

func Test__When_root_spec_fails_sporadically__Then_the_errors_are_merged(t *testing.T) {
	runner := NewRunner()
	runner.AddSpec("RootSpec", func(c Context) {
		i := 0
		c.Specify("Child A", func() {
			i = 1
		})
		c.Specify("Child B", func() {
			i = 2
		})
		c.Then(10).Should.Equal(20)     // stays same - will be reported once
		c.Then(10 + i).Should.Equal(20) // changes - will be reported many times
	})
	runner.Run()

	assertReportIs(runner.Results(), `
- RootSpec [FAIL]
    Expected '20' but was '10'
    Expected '20' but was '11'
    Expected '20' but was '12'
  - Child A
  - Child B

3 specs, 1 failures
`,
		t)
}

func Test__When_non_root_spec_fails_sporadically__Then_the_errors_are_merged(t *testing.T) {
	runner := NewRunner()
	runner.AddSpec("RootSpec", func(c Context) {
		c.Specify("Failing", func() {
			i := 0
			c.Specify("Child A", func() {
				i = 1
			})
			c.Specify("Child B", func() {
				i = 2
			})
			c.Then(10).Should.Equal(20)     // stays same - will be reported once
			c.Then(10 + i).Should.Equal(20) // changes - will be reported many times
		})
	})
	runner.Run()

	assertReportIs(runner.Results(), `
- RootSpec
  - Failing [FAIL]
      Expected '20' but was '10'
      Expected '20' but was '11'
      Expected '20' but was '12'
    - Child A
    - Child B

4 specs, 1 failures
`,
		t)
}

func assertReportIs(results *ResultCollector, expected string, t *testing.T) {
	out := new(bytes.Buffer)
	results.Visit(NewPrinter(SimplePrintFormat(out)))
	report := out.String()
	assertEqualsTrim(expected, report, t)
}
