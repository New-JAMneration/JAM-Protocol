package cli

import (
	"flag"
	"fmt"
)

type Float64Flag struct {
	Name         string
	Usage        string
	DefaultValue float64
	Destination  *float64
}

func (f *Float64Flag) Apply(set *flag.FlagSet) {
	set.Float64Var(f.Destination, f.Name, f.DefaultValue, f.Usage)
}

func (f *Float64Flag) String() string {
	return fmt.Sprintf("float64  %s", f.Usage)
}

func (f *Float64Flag) NameStr() string {
	return f.Name
}
