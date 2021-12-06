package main

import (
	"fmt"
	"encoding/json"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/interpreter/functions"
	"google.golang.org/genproto/googleapis/api/expr/v1alpha1"

	"github.com/blang/semver/v4"
)

type semverLib struct{}

func (semverLib) CompileOptions() []cel.EnvOption {
	return []cel.EnvOption{
		cel.Declarations(
			decls.NewFunction("semver_compare",
				decls.NewOverload("semver_compare",
					[]*expr.Type{decls.Any, decls.Any},
					decls.Int))),
	}
}

func (semverLib) ProgramOptions() []cel.ProgramOption {
	return []cel.ProgramOption{
		cel.Functions(
			&functions.Overload{
				Operator: "semver_compare",
				Binary:   semverCompare,
			},
		),
	}
}

func main() {
	env, err := cel.NewEnv(cel.Lib(semverLib{}), cel.Declarations(
		decls.NewVar("ocpversion", decls.String), decls.NewIdent("properties", decls.NewMapType(decls.Any, decls.Any),nil)))

	if err != nil {
		fmt.Printf("new env error: %s", err)
	}

	string1 := `"4.9"`

	string2 := `
	{
	  "group": "a2",
	  "version": "b",
	  "kind": "c"
	}`

	var v interface{}
	if err := json.Unmarshal([]byte(string1), &v); err != nil {
		panic(err)
	}

	props := make([]map[string]interface{}, 2)
	props[0] = map[string]interface{}{
	"type":  "olm.maxOpenShiftVersion",
	"value": v,
	}

	if err := json.Unmarshal([]byte(string2), &v); err != nil {
		panic(err)
	}

	props[1] = map[string]interface{}{
	"type":  "olm.gvk",
	"value": v,
	}

	newpros := map[string]interface{}{"properties": props}

	// Parse and check the expression.
	p, issues := env.Parse("properties.exists(p, p.type == 'olm.maxOpenShiftVersion' && (semver_compare(p.value, 4.8) <= 0))")

	//p, issues := env.Parse("properties.exists(p, p.type == 'olm.maxOpenShiftVersion' && p.value == {'group': 'a1', 'version': 'b', 'kind': 'c'})")
	if issues != nil && issues.Err() != nil {
		fmt.Printf("parse error: %s", issues.Err())
	}

	c, issues := env.Check(p)
	if issues != nil && issues.Err() != nil {
		fmt.Printf("check error: %s", issues.Err())
	}

	// Evaluate the program against some inputs.
	prg, err := env.Program(c)
	if err != nil {
		fmt.Printf("program error: %s", err.Error())
	}

	out, _, err := prg.Eval(newpros)
	if err != nil {
		fmt.Printf("eval error: %s", err.Error())
	}

	// Result: true
	fmt.Println(out)
}

func semverCompare(val1, val2 ref.Val) ref.Val {
	// str, ok := val1.(types.String)
	// if !ok {
	// 	return types.ValOrErr(str, "unexpected type '%v'", val1.Type())
	// }
	//
	// str2, ok := val2.(types.String)
	// if !ok {
	// 	return types.ValOrErr(str, "unexpected type '%v'", val2.Type())
	// }

	v1, err := semver.ParseTolerant(fmt.Sprint(val1.Value()))
	if err != nil {
		return types.ValOrErr(val1, "unable to parse '%v' to semver format", val1.Value())
	}

	v2, err := semver.ParseTolerant(fmt.Sprint(val2.Value()))
	if err != nil {
		return types.ValOrErr(val2, "unable to parse '%v' to semver format", val2.Value())
	}

	return types.Int(v1.Compare(v2))
}
