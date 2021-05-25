package dagger

import (
	"context"

	"cuelang.org/go/cue"
	"dagger.io/go/dagger/compiler"
	"github.com/rs/zerolog/log"
)

func isReference(val cue.Value) bool {
	isRef := func(v cue.Value) bool {
		_, ref := v.ReferencePath()

		if ref.String() == "" || v.Path().String() == ref.String() {
			// not a reference
			return false
		}

		for _, s := range ref.Selectors() {
			if s.IsDefinition() {
				// if we reference to a definition, we skip the check
				return false
			}
		}

		return true
	}

	op, vals := val.Expr()
	if op == cue.NoOp {
		return isRef(val)
	}

	for _, v := range vals {
		// if the expr has an op (& or |, etc...), check the expr values, recursively
		if isReference(v) {
			return true
		}
	}

	return isRef(val)
}

func ScanInputs(ctx context.Context, value *compiler.Value) []*compiler.Value {
	lg := log.Ctx(ctx)
	inputs := []*compiler.Value{}

	value.Walk(
		func(val *compiler.Value) bool {
			if isReference(val.Cue()) {
				lg.Debug().Str("value.Path", val.Path().String()).Msg("found reference, stop walk")
				return false
			}

			if !val.HasAttr("input") {
				return true
			}

			lg.Debug().Str("value.Path", val.Path().String()).Msg("found input")
			inputs = append(inputs, val)

			return true
		}, nil,
	)

	return inputs
}