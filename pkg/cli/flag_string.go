package cli

import (
	"flag"
	"fmt"
)

type StringFlag struct {
	Name         string
	Usage        string
	DefaultValue string
	Destination  *string
}

func (f *StringFlag) Apply(set *flag.FlagSet) {
	set.StringVar(f.Destination, f.Name, f.DefaultValue, f.Usage)
}

func (f *StringFlag) String() string {
	return fmt.Sprintf("string  %s", f.Usage)
}

func (f *StringFlag) NameStr() string {
	return f.Name
}
