package cli

import (
	"flag"
	"testing"
)

func TestBoolFlag(t *testing.T) {
	var test bool

	f := &BooleanFlag{
		Name:        "bool",
		Usage:       "test bool usage",
		Destination: &test,
	}

	set := flag.NewFlagSet("test", 0)
	f.Apply(set)

	err := set.Parse([]string{"--bool"})
	if err != nil {
		t.Fatal(err)
	}

	if !test {
		t.Fatal("Should be true")
	}
}

func TestStringFlag(t *testing.T) {
	var test string

	f := &StringFlag{
		Name:        "str",
		Usage:       "test str usage",
		Destination: &test,
	}

	set := flag.NewFlagSet("test", 0)
	f.Apply(set)

	err := set.Parse([]string{"--str", "hello world"})
	if err != nil {
		t.Fatal(err)
	}

	if test != "hello world" {
		t.Fatal("Should be \"hello world\"")
	}
}

func TestIntFlag(t *testing.T) {
	var test int

	f := &IntFlag{
		Name:        "int",
		Usage:       "test int usage",
		Destination: &test,
	}

	set := flag.NewFlagSet("test", 0)
	f.Apply(set)

	err := set.Parse([]string{"--int", "1234"})
	if err != nil {
		t.Fatal(err)
	}

	if test != 1234 {
		t.Fatal("Should be \"1234\"")
	}
}

func TestInt64Flag(t *testing.T) {
	var test int64

	f := &Int64Flag{
		Name:        "int64",
		Usage:       "test int64 usage",
		Destination: &test,
	}

	set := flag.NewFlagSet("test", flag.ExitOnError)
	f.Apply(set)

	err := set.Parse([]string{"--int64", "1234"})
	if err != nil {
		t.Fatal(err)
	}

	if test != 1234 {
		t.Fatal("Should be \"1234\"")
	}
}

func TestFloat64Flag(t *testing.T) {
	var test float64

	f := &Float64Flag{
		Name:        "float64",
		Usage:       "test float64 usage",
		Destination: &test,
	}

	set := flag.NewFlagSet("test", flag.ExitOnError)
	f.Apply(set)

	err := set.Parse([]string{"--float64", "1234.125"})
	if err != nil {
		t.Fatal(err)
	}

	if test != 1234.125 {
		t.Fatal("Should be \"1234.125\"")
	}
}

func TestMultipleFlag(t *testing.T) {
	var str string
	var i int
	var i64 int64
	var f64 float64
	var b bool

	flags := []Flag{
		&StringFlag{
			Name:        "str",
			Usage:       "test str usage",
			Destination: &str,
		},
		&IntFlag{
			Name:        "int",
			Usage:       "test int usage",
			Destination: &i,
		},
		&Int64Flag{
			Name:        "int64",
			Usage:       "test int64 usage",
			Destination: &i64,
		},
		&Float64Flag{
			Name:        "float64",
			Usage:       "test float64 usage",
			Destination: &f64,
		},
		&BooleanFlag{
			Name:        "boolean",
			Usage:       "test boolean usage",
			Destination: &b,
		},
	}

	set := flag.NewFlagSet("test", flag.ExitOnError)

	for _, baseFlag := range flags {
		baseFlag.Apply(set)
	}

	err := set.Parse([]string{"--boolean", "--str", "1234", "--int", "1234", "--int64", "1234", "--float64", "1234.125"})
	if err != nil {
		t.Fatal(err)
	}

	if str != "1234" {
		t.Fatal("Should be \"1234\"")
	}

	if i != 1234 {
		t.Fatal("Should be \"1234\"")
	}

	if i64 != 1234 {
		t.Fatal("Should be \"1234\"")
	}

	if f64 != 1234.125 {
		t.Fatal("Should be \"1234.125\"")
	}

	if !b {
		t.Fatal("Should be true")
	}
}
