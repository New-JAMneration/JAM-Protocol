package extrinsic

import (
	"reflect"
	"testing"
)

func TestNewAssurance(t *testing.T) {
	type args struct {
		Anchor         string
		Bitstring      string
		ValidatorIndex uint16
		Signature      string
	}
	tests := []struct {
		name string
		args Assurance
		want *Assurance
	}{
		{
			name: "Test case 1",
			args: Assurance{
				Anchor:         "0x0cffbf67aae50aeed3c6f8f0d9bf7d854ffd87cef8358cbbaa587a9e3bd1a776",
				Bitstring:      "0x01",
				ValidatorIndex: 0,
				Signature:      "0x2d8ec7b235be3b3cbe9be3d5ff36f082942102d64a0dc5953709a95cca55b58b1af297f534d464264be77477b547f3c596b947edbca33f6631f1aa188d25a38b",
			},
			want: &Assurance{
				Anchor:         "0x0cffbf67aae50aeed3c6f8f0d9bf7d854ffd87cef8358cbbaa587a9e3bd1a776",
				Bitstring:      "0x01",
				ValidatorIndex: 0,
				Signature:      "0x2d8ec7b235be3b3cbe9be3d5ff36f082942102d64a0dc5953709a95cca55b58b1af297f534d464264be77477b547f3c596b947edbca33f6631f1aa188d25a38b",
			},
		},
		{
			name: "Test case 2",
			args: Assurance{
				Anchor:         "0x2398ce69c3585e1b1b574a5a7185a2a086350abd4606d15aace8b4610b494772",
				Bitstring:      "0x01",
				ValidatorIndex: 1,
				Signature:      "0xdda7a577f150ee83afedc9d3b50a4f00fcf21248e6f73097abcc4bb634f854aedc53769838d294b09c0184fb0e66f09bae8cc243f842a6cc401488591e9ffdb1",
			},
			want: &Assurance{
				Anchor:         "0x2398ce69c3585e1b1b574a5a7185a2a086350abd4606d15aace8b4610b494772",
				Bitstring:      "0x01",
				ValidatorIndex: 1,
				Signature:      "0xdda7a577f150ee83afedc9d3b50a4f00fcf21248e6f73097abcc4bb634f854aedc53769838d294b09c0184fb0e66f09bae8cc243f842a6cc401488591e9ffdb1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewAssurance(tt.args.Anchor, tt.args.Bitstring, tt.args.ValidatorIndex, tt.args.Signature); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewAssurance() = %v, want %v", got, tt.want)
			}
		})
	}
}
