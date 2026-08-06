package PVM

// Host-function gas costs from Gray Paper definitions.tex (Host-function gas costs,
// lines 302–367). Comments use PDF notation: M_□ for rates, 𝒢(L, ℓ) = ⌈L·ℓ/1024⌉
// (eq:fnmemgas; implemented as MemGas). Order matches:
// https://github.com/gavofyork/graypaper/blob/72bee14497387d43d8ba465436efe66b281982c1/text/definitions.tex
const (
	HostGasUnknown Gas = 1000 // M_∅ — Gas cost charged for an unknown host-call.

	HostGasAssign Gas = 1818 // M_A — Ω_A (assign) base gas cost.

	HostGasBlessConst Gas = 422 // M_{B,c} — Ω_B (bless) base gas cost.
	HostGasBlessItem  Gas = 20  // M_{B,ℓ} — Ω_B (bless) gas per item.

	HostGasCheckpoint Gas = 103 // M_C — Ω_C (checkpoint) base gas cost.

	HostGasDesignateConst     Gas = 1100 // M_{D,c} — Ω_D (designate) base gas cost.
	HostGasDesignateValidator Gas = 302  // M_{D,ℓ} — Ω_D (designate) gas per validator.

	HostGasExport Gas = 3521 // M_E — Ω_E (export) base gas cost.

	HostGasForget Gas = 3250 // M_F — Ω_F (forget) base gas cost.

	HostGasGas Gas = 48 // M_G — Ω_G (gas) base gas cost.

	HostGasHistoricalLookupConst  Gas = 1125 // M_{H,c} — Ω_H (historical_lookup) base gas cost.
	HostGasHistoricalLookupOctets Gas = 264  // M_{H,ℓ} — Ω_H (historical_lookup) gas per 1024 octets (𝒢(M_{H,ℓ}, ℓ)).

	HostGasInfo Gas = 703 // M_I — Ω_I (info) base gas cost.

	HostGasEject Gas = 458 // M_J — Ω_J (eject) base gas cost.

	HostGasInvoke Gas = 968 // M_K — Ω_K (invoke) base gas cost.

	HostGasLookupConst  Gas = 600 // M_{L,c} — Ω_L (lookup) base gas cost.
	HostGasLookupOctets Gas = 248 // M_{L,ℓ} — Ω_L (lookup) gas per 1024 octets (𝒢(M_{L,ℓ}, ℓ)).

	HostGasMachineConst  Gas = 1862 // M_{M,c} — Ω_M (machine) base gas cost.
	HostGasMachineOctets Gas = 112  // M_{M,ℓ} — Ω_M (machine) gas per 1024 octets, program size (𝒢(M_{M,ℓ}, ℓ)).

	HostGasNew Gas = 3855 // M_N — Ω_N (new) base gas cost.

	HostGasPokeConst  Gas = 297 // M_{O,c} — Ω_O (poke) base gas cost.
	HostGasPokeOctets Gas = 224 // M_{O,ℓ} — Ω_O (poke) gas per 1024 octets (𝒢(M_{O,ℓ}, ℓ)).

	HostGasPeekConst  Gas = 377 // M_{P,c} — Ω_P (peek) base gas cost.
	HostGasPeekOctets Gas = 336 // M_{P,ℓ} — Ω_P (peek) gas per 1024 octets (𝒢(M_{P,ℓ}, ℓ)).

	HostGasQuery Gas = 643 // M_Q — Ω_Q (query) base gas cost.

	HostGasReadConst       Gas = 2407 // M_{R,c} — Ω_R (read) base gas cost.
	HostGasReadKeyOctets   Gas = 1736 // M_{R,k,ℓ} — Ω_R (read) key gas per 1024 octets (𝒢(M_{R,k,ℓ}, ℓ)).
	HostGasReadValueOctets Gas = 248  // M_{R,v,ℓ} — Ω_R (read) value gas per 1024 octets (𝒢(M_{R,v,ℓ}, ℓ)).

	HostGasSolicit Gas = 2193 // M_S — Ω_S (solicit) base gas cost.

	HostGasTransfer Gas = 575 // M_T — Ω_T (transfer) base gas cost.

	HostGasUpgrade Gas = 1028 // M_U — Ω_U (upgrade) base gas cost.

	HostGasWriteConst     Gas = 2442 // M_{W,c} — Ω_W (write) base gas cost.
	HostGasWriteValOctets Gas = 216  // M_{W,v,ℓ} — Ω_W (write) gas per 1024 octets (𝒢(M_{W,v,ℓ}, ℓ)).
	HostGasWriteKeyOctets Gas = 3358 // M_{W,k,ℓ} — Ω_W (write) key gas per 1024 octets (𝒢(M_{W,k,ℓ}, ℓ)).

	HostGasExpunge Gas = 335 // M_X — Ω_X (expunge) base gas cost.
)

// fetchGasCosts holds M_{Y,i,c} / M_{Y,i,ℓ} per fetch discriminator i (definitions.tex lines 337–352).
var fetchGasCosts = [16]struct {
	constant Gas
	octets   Gas
}{
	0:  {390, 0},   // M_{Y,0,c}, M_{Y,0,ℓ} — Ω_Y (fetch) case 0: protocol parameters.
	1:  {103, 0},   // M_{Y,1,c}, M_{Y,1,ℓ} — Ω_Y (fetch) case 1: entropy.
	2:  {80, 96},   // M_{Y,2,c}, M_{Y,2,ℓ} — Ω_Y (fetch) case 2: auth trace.
	3:  {85, 96},   // M_{Y,3,c}, M_{Y,3,ℓ} — Ω_Y (fetch) case 3: any extrinsic, by index.
	4:  {85, 96},   // M_{Y,4,c}, M_{Y,4,ℓ} — Ω_Y (fetch) case 4: our extrinsic, by index.
	5:  {171, 0},   // M_{Y,5,c}, M_{Y,5,ℓ} — Ω_Y (fetch) case 5: any import, by index.
	6:  {171, 0},   // M_{Y,6,c}, M_{Y,6,ℓ} — Ω_Y (fetch) case 6: our import, by index.
	7:  {85, 96},   // M_{Y,7,c}, M_{Y,7,ℓ} — Ω_Y (fetch) case 7: encoded work-package.
	8:  {84, 0},    // M_{Y,8,c}, M_{Y,8,ℓ} — Ω_Y (fetch) case 8: auth config.
	9:  {88, 0},    // M_{Y,9,c}, M_{Y,9,ℓ} — Ω_Y (fetch) case 9: auth token.
	10: {111, 0},   // M_{Y,10,c}, M_{Y,10,ℓ} — Ω_Y (fetch) case 10: refine context.
	11: {317, 0},   // M_{Y,11,c}, M_{Y,11,ℓ} — Ω_Y (fetch) case 11: items summary.
	12: {250, 0},   // M_{Y,12,c}, M_{Y,12,ℓ} — Ω_Y (fetch) case 12: any item summary.
	13: {95, 96},   // M_{Y,13,c}, M_{Y,13,ℓ} — Ω_Y (fetch) case 13: any payload.
	14: {287, 400}, // M_{Y,14,c}, M_{Y,14,ℓ} — Ω_Y (fetch) case 14: accumulate items.
	15: {355, 344}, // M_{Y,15,c}, M_{Y,15,ℓ} — Ω_Y (fetch) case 15: any accumulate item.
}

const (
	HostGasFetchOtherConst  Gas = 80 // M_{Y,∅,c} — Ω_Y (fetch) otherwise: nothing.
	HostGasFetchOtherOctets Gas = 0  // M_{Y,∅,ℓ}

	HostGasPagesAllocConst   Gas = 275 // M_{Z,a,c} — Ω_Z (pages) alloc base gas cost.
	HostGasPagesAllocPage    Gas = 121 // M_{Z,a,p} — Ω_Z (pages) alloc gas per page.
	HostGasPagesFreeConst    Gas = 212 // M_{Z,f,c} — Ω_Z (pages) free base gas cost.
	HostGasPagesFreePage     Gas = 118 // M_{Z,f,p} — Ω_Z (pages) free gas per page.
	HostGasPagesSetModeConst Gas = 130 // M_{Z,s,c} — Ω_Z (pages) setmode base gas cost.
	HostGasPagesSetModePage  Gas = 29  // M_{Z,s,p} — Ω_Z (pages) setmode gas per page.
	HostGasPagesInvalid      Gas = 80  // M_{Z,i} — Ω_Z (pages) invalid-r base gas cost.

	HostGasYield Gas = 98 // M_♉ — Ω_♉ (yield) base gas cost.

	HostGasProvideConst  Gas = 3980 // M_{♈,c} — Ω_♈ (provide) base gas cost.
	HostGasProvideOctets Gas = 2264 // M_{♈,ℓ} — Ω_♈ (provide) gas per 1024 octets (𝒢(M_{♈,ℓ}, ℓ)).

	HostGasGrowHeapConst Gas = 275 // M_{♊,c} — Ω_♊ (grow_heap) base gas cost.
	HostGasGrowHeapPage  Gas = 121 // M_{♊,p} — Ω_♊ (grow_heap) gas per additional page.
)

// Non–Gray Paper host-call gas (JIP-1 extensions).
const (
	HostGasLog Gas = 10 // Ω_log (opcode 100) — fixed cost; not defined in GP.
)

// FetchGasCost returns M_{Y,i,c} and M_{Y,i,ℓ} (the L in 𝒢(L, ℓ)) for fetch Ω_Y
// discriminator i, falling back to M_{Y,∅,*} for undefined cases.
func FetchGasCost(discriminator uint64) (constant, octets Gas) {
	if discriminator >= uint64(len(fetchGasCosts)) {
		return HostGasFetchOtherConst, HostGasFetchOtherOctets
	}
	c := fetchGasCosts[discriminator]
	return c.constant, c.octets
}
