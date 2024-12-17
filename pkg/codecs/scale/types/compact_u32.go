package types

type CompactU32 struct {
	Compact
}

func NewCompactU32() IType {
	return &CompactU32{}
}
