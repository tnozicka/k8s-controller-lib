# Agent instructions

This file provides guidance to AI agents when working with code in this repository.

## Working with the repository

### Quick verification that everything compiles
```bash
go vet ./...
```

### Running tools
Most tools are vendored and should be run using `go tool`.
`go.mod` has a tools section containing the list of tools available this way.

## Dependency update
Follow the instructions below to update dependencies for specific dependency managers.
Alway run `go vet ./...` after updating dependencies to make sure it compiles.

### Go
```bash
go get -u ./...
go mod tidy
go mod vendor
```

# Mandatory instructions
You are a helpful coding agent that shall strictly adhere to the following rules:
- You shall not lie. If you don't know, say so. If you don't feel confident, require extra input.
- Focus on code reusability and maintainability.
- Do not ignore return values.
- When you are asked to do a change in the whole project and the whole code base, you need to inspect every single package and file and you shall verify that all of those files confirm to the requirements, after you make those changes.
- Prefer using the standard library over third party packages.
- Prefer using k8s.io libraries.
- Do not edit code in `/thirdparty` directory.
- For any change you make, prefer following the existing patterns found in this repository. 
- /internal/* is not a good example, unless you are inside that path.
- Prefer minimal changes to existing code.
- Do not use deprecated functions. Always check whether the function you want to use is deprecated.
- Do not add new comments unless you are copying an existing pattern.
- Never edit `go.mod`, `go.sum`, `vendor/`, `_vendor` directly.
- Anything that `make update` changes is a generated file that you should not edit directly.
- Use CamelCase for Go. All characters of an abbreviation shall have the same capitalization.
- Name imports consistently with the code base. Follow the most frequent naming pattern in the project.
- Run `make update-gofmt` after every change of a Go file.
- Make sure every line of code you add is actually used.
- In the same file, a function should be first defined, then used.
- Do not ignore tool failures. No failures are preexisting. `git` is allowed not to work, but you don't need it.
- Do not run any `git` commands.
- Use original type names and capitalization (without extra spaces) in error messages.
- Do not shadow variables.
- Always split variable assignment and if condition into separate statements.
- Prefer FlagSet variants with shorthand even if empty, so it's consistent with other flags.
- Tests need to be minimal, meaningful, non-overlapping and increase test coverage. You should iterate to find a minimal set of test cases that produce the largest code coverage.
- Avoid creating global variables.
- Prefer abstractions and sharing configuration over code duplication.
- Write robust code, handle all logic cases explicitly. Error out when inputs are incorrect.
- Don't replace listers with live client calls.
- Use the new() function to create pointer types as of go 1.26
- Split function arguments or initializers to new lines when the line is longer than 140 characters.
- Avoid extra nesting by preferring the main condition branch to end the flow (e.g. with return or continue) and avoiding the need to an else block, where possible.
- Error messages that wrap other messages or errors should follow the format: `can't <verb> ...`
- Prefer consistent order, like when creating variables that are later input to a function, it should be in the same order as in the function signature.
- `go list` is more reliable than `grep`. Use it to learn about functions, symbols, packages, modules and so on.
- When you are done run `go vet` to make sure it all compiles and then `go fix` on the code that you've changed. Disregard formatting changes on the code lines not changed by you.
- Initialize structures with one line per field.
- If function arguments need to be split to multiple lines, use a dedicated line for each of them and the braces.
- Prefer wait.Group when running tasks in the background. Always wait for the group to finish.
- Don't look up unknown functions in the vendor folder directly. Go only vendors the packages (folders) that are currently used and not the rest.

## Go unit tests

### Generic instructions
- Tests for a function belong to the file with the same name and `_test.go` suffix.
- Use `apiequality.Semantic.DeepEqual` for Kubernetes types.
- `cmp.Diff` can only be used to format error strings
- Expected values in tests should be inlined and not rely on external functions.
- Don't use DeepEqual for Go native types. Interfaces like errors should use DeepEqual.
- `t.Errorf` and `t.Fatalf` should follow the format for error messages.
- When comparing the `got` use compoud tpye for the expected value and rely on the DeepEqual. Do not compare fields separately.
- Helpers in unit tests should be marked with `t.Helper()`.
- Test use t.Context().
- Limit unit test creation to the test functions. Don't create other functions or symbols unless you are instructed to. 
- Do not create your own fakes or mocks. Try to understand existing patterns in all other tests and how you could use a similar pattern.


### Table driven
This is the pattern for table-driven tests.
All generic instructions apply. 
Most tests should be table-driven.
```
t.Parallel()

tt := []struct {
	name     string
	argFoo string
	argBar string
	expected string
	expectedErr error
}{
	// ...
}

for _, tc := range tt {	 
	t.Run(tc.name, func(t *testing.T) {
		t.Parallel()

		got, err := TestFunction(tc.argFoo, tc.argBar)
		if !reflect.DeepEqual(err, tc.expectedErr) {
			t.Errorf("expected and got errors differ:\n%s", cmp.Diff(tc.expectedErr, err))
		}
		
		if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("expected and got differ:\n%s", cmp.Diff(tc.expected, got))
		}
	})
```

- Test case names should be minimal and not ambiguous.
- The test table (tt) items should only contain the test name, arguments and expected values. Expected values that are compared with the results from the function being tested shall be prefixed with `expected` followed by the result name.
- Arguments should be included in the test case structure only if they have a distinct value. E.g. fakes may be identical for all test cases and should be initialized withing the subtest, just before the tested function is called. 
- In the test table initialization, all variables should be explicitly initialized even with zero values.
- Order the function arguments in the test table in the same order as they are in the function signature.
- Use slices instead of maps for expected and input objects.
- When testing a generic function taking an interface, use the same interface in the test table for expected and input objects.
- In the test table, there should be one item of the same type as the tested function return value, for each of them.
- Tests shouldn't sort the test cases at runtime but have them already sorted in the test table.
