package cli

import "flag"

type Flag interface {
	Apply(*flag.FlagSet)
	String() string
	NameStr() string
}
