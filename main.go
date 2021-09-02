package main

import (
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/interpreter/functions"
	"google.golang.org/genproto/googleapis/api/expr/v1alpha1"

	"github.com/blang/semver/v4"
)

func main() {
	dec := cel.Declarations(
	// Identifiers used within this expression.
	decls.NewVar("ocpversion", decls.String),
	// Function to generate a greeting from one person to another.
	//    i.greet(you)
	decls.NewFunction("semver_compare",
		decls.NewOverload("semver_compare",
			[]*expr.Type{decls.String, decls.String},
			decls.Int)))

	e, _ := cel.NewEnv(dec)

	// Parse and check the expression.
	p, issues := e.Parse("semver_compare(ocpversion, '4.8.0') == 1")
	if issues != nil && issues.Err() != nil {
		fmt.Printf("parse error: %s", issues.Err())
	}

	c, issues := e.Check(p)
	if issues != nil && issues.Err() != nil {
		fmt.Printf("check error: %s", issues.Err())
	}

	funcs := cel.Functions(
		&functions.Overload{
			Operator: "semver_compare",
			Binary:   SemverCompare,
		})

	// Evaluate the program against some inputs.
	prg, err := e.Program(c, funcs)
	if err != nil {
		fmt.Printf("program error: %s", err.Error())
	}

	out, _, err := prg.Eval(map[string]interface{}{"ocpversion": "4.9.0"})
	if err != nil {
		fmt.Printf("eval error: %s", err.Error())
	}

	// Result: true
	fmt.Println(out)
}

func SemverCompare(val1, val2 ref.Val) ref.Val {
	str, ok := val1.(types.String)
	if !ok {
		return types.ValOrErr(str, "unexpected type '%v'", val1.Type())
	}

	str2, ok := val2.(types.String)
	if !ok {
		return types.ValOrErr(str, "unexpected type '%v'", val2.Type())
	}

	v1, err := semver.ParseTolerant(string(str))
	if err != nil {
		return types.ValOrErr(str, "unable to parse '%v' to semver format", val1.Value())
	}

	v2, err := semver.ParseTolerant(string(str2))
	if err != nil {
		return types.ValOrErr(str, "unable to parse '%v' to semver format", val2.Value())
	}

	result := v1.Compare(v2)

	return types.Int(result)
}
