package cli

import (
	"flag"
	"fmt"
)

type IntFlag struct {
	Name         string
	Usage        string
	DefaultValue int
	Destination  *int
}

func (f *IntFlag) Apply(set *flag.FlagSet) {
	set.IntVar(f.Destination, f.Name, f.DefaultValue, f.Usage)
}

func (f *IntFlag) String() string {
	return fmt.Sprintf("int  %s", f.Usage)
}

func (f *IntFlag) NameStr() string {
	return f.Name
}
