package cli

import (
	"flag"
	"fmt"
)

type BooleanFlag struct {
	Name         string
	Usage        string
	DefaultValue bool
	Destination  *bool
}

func (f *BooleanFlag) Apply(set *flag.FlagSet) {
	set.BoolVar(f.Destination, f.Name, f.DefaultValue, f.Usage)
}

func (f *BooleanFlag) String() string {
	return fmt.Sprintf("boolean  %s", f.Usage)
}

func (f *BooleanFlag) NameStr() string {
	return f.Name
}
