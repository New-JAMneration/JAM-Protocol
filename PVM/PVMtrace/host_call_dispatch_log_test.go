package PVMtrace

import "testing"

func TestParseHostCallDispatchLogEnv(t *testing.T) {
	tests := []struct {
		name string
		val  string
		want bool
	}{
		{name: "empty", val: "", want: false},
		{name: "unset-whitespace", val: "  ", want: false},
		{name: "zero", val: "0", want: false},
		{name: "false", val: "false", want: false},
		{name: "one", val: "1", want: true},
		{name: "true", val: "true", want: true},
		{name: "TRUE", val: "TRUE", want: true},
		{name: "yes", val: "yes", want: true},
		{name: "YES", val: "YES", want: true},
		{name: "on", val: "on", want: true},
		{name: "ON", val: "ON", want: true},
		{name: "padded-one", val: " 1 ", want: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := parseHostCallDispatchLogEnv(tc.val); got != tc.want {
				t.Fatalf("parseHostCallDispatchLogEnv(%q)=%v, want %v", tc.val, got, tc.want)
			}
		})
	}
}

func TestHostCallDispatchLogEnabledIgnoresRuntimeSetenv(t *testing.T) {
	enabled := HostCallDispatchLogEnabled()
	flip := "1"
	if enabled {
		flip = "0"
	}
	t.Setenv(envHostCallLog, flip)
	if HostCallDispatchLogEnabled() != enabled {
		t.Fatal("JAM_PVM_HOSTCALL_LOG is snapshotted at init; runtime env changes must not apply")
	}
}

func BenchmarkLogHostCallDispatchEnvDisabled(b *testing.B) {
	b.ReportAllocs()
	regs := [13]uint64{}
	for b.Loop() {
		if HostCallDispatchLogEnabled() {
			LogHostCallDispatchEnv(-1, "fetch", regs, 0)
		}
	}
}
