package semantic

import (
	"fmt"
)

type InferType struct {
	ID       int
	Resolved Type
}

func (i *InferType) Signature() TypeSignature {
	if i.Resolved == nil {
		return TypeSignature{
			BaseType: fmt.Sprintf("_infer T%d", i.ID),
		}
	} else {
		return i.Resolved.Signature()
	}
}

func (i *InferType) String() string {
	return i.Signature().String()
}

func Resolve(t Type) Type {
	if t == nil {
		return nil
	}
	infer, ok := t.(*InferType)
	if !ok {
		return t
	}

	if infer.Resolved == nil {
		return infer
	}

	infer.Resolved = Resolve(infer.Resolved)
	return infer.Resolved
}
