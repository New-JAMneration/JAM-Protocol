package cli

import (
	"flag"
	"fmt"
)

type Int64Flag struct {
	Name         string
	Usage        string
	DefaultValue int64
	Destination  *int64
}

func (f *Int64Flag) Apply(set *flag.FlagSet) {
	set.Int64Var(f.Destination, f.Name, f.DefaultValue, f.Usage)
}

func (f *Int64Flag) String() string {
	return fmt.Sprintf("int64  %s", f.Usage)
}

func (f *Int64Flag) NameStr() string {
	return f.Name
}
