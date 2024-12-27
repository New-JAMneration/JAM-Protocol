package types

import "github.com/New-JAMneration/JAM-Protocol/pkg/codecs/scale/types"

func InitScaleRegistry() {
	m := map[string]func() types.IType{
		// sample
		"bytesequence": types.NewHexBytes,
		"opaquehash":   NewOpaqueHash,
		// crypto
		"bandersnatchpublic":           func() types.IType { return NewByteArray(32) },
		"ed25519public":                func() types.IType { return NewByteArray(32) },
		"ed25519signature":             func() types.IType { return NewByteArray(64) },
		"blspublic":                    func() types.IType { return NewByteArray(144) },
		"bandersnatchvrfsignature":     func() types.IType { return NewByteArray(96) },
		"bandersnatchringvrfsignature": func() types.IType { return NewByteArray(784) },
		"bandersnatchringcommitment":   func() types.IType { return NewByteArray(144) },

		// Application Specific Core
		"timeslot":          types.NewU32,
		"epochindex":        types.NewU32,
		"validatorindex":    types.NewU16,
		"coreindex":         types.NewU16,
		"headerhash":        NewOpaqueHash,
		"stateroot":         NewOpaqueHash,
		"beefyroot":         NewOpaqueHash,
		"workpackagehash":   NewOpaqueHash,
		"workreporthash":    NewOpaqueHash,
		"exportsroot":       NewOpaqueHash,
		"erasureroot":       NewOpaqueHash,
		"gas":               types.NewU64,
		"entropy":           NewOpaqueHash,
		"entropybuffer":     func() types.IType { return types.NewFixedArray(4, "u8") },
		"validatormetadata": func() types.IType { return NewByteArray(128) },
		"validator": func() types.IType {
			maps := []types.TypeMap{
				{Name: "bandersnatch", Type: "bandersnatchpublic"},
				{Name: "ed25519", Type: "ed25519public"},
				{Name: "bls", Type: "blspublic"},
				{Name: "metadata", Type: "validatormetadata"},
			}
			return types.NewStruct(maps)
		},
		"validatorsdata": func() types.IType { return types.NewFixedArray(ValidatorsCount, "validatordata") },

		// Service

		"serviceid": types.NewU32,
		"serviceinfo": func() types.IType {
			maps := []types.TypeMap{
				{Name: "code_hash", Type: "opaquehash"},
				{Name: "balance", Type: "u64"},
				{Name: "min_item_gas", Type: "gas"},
				{Name: "min_memo_gas", Type: "gas"},
				{Name: "bytes", Type: "u64"},
				{Name: "items", Type: "u32"},
			}
			return types.NewStruct(maps)
		},

		// Availability Assignments
		"availabilityassignment": func() types.IType {
			maps := []types.TypeMap{
				{Name: "report", Type: "workreport"},
				{Name: "timeout", Type: "u32"},
			}
			return types.NewStruct(maps)
		},
		"availabilityassignments": func() types.IType { return types.NewFixedArray(CoresCount, "option<availabilityassignment>") },

		// Refine Context
		"refinecontext": func() types.IType {
			maps := []types.TypeMap{
				{Name: "anchor", Type: "headerhash"},
				{Name: "state_root", Type: "stateroot"},
				{Name: "beefy_root", Type: "beefyroot"},
				{Name: "lookup_anchor", Type: "headerhash"},
				{Name: "lookup_anchor_slot", Type: "timeslot"},
				{Name: "prerequisites", Type: "vec<opaquehash>"},
			}
			return types.NewStruct(maps)
		},

		// Work Package
		"importspec": func() types.IType {
			maps := []types.TypeMap{
				{Name: "tree_root", Type: "opaquehash"},
				{Name: "index", Type: "u16"},
			}
			return types.NewStruct(maps)
		},
		"extrinsicspec": func() types.IType {
			maps := []types.TypeMap{
				{Name: "hash", Type: "opaquehash"},
				{Name: "len", Type: "u32"},
			}
			return types.NewStruct(maps)
		},
		"authorizer": func() types.IType {
			maps := []types.TypeMap{
				{Name: "code_hash", Type: "opaquehash"},
				{Name: "params", Type: "bytesequence"},
			}
			return types.NewStruct(maps)
		},
		"workitem": func() types.IType {
			maps := []types.TypeMap{
				{Name: "service", Type: "serviceid"},
				{Name: "code_hash", Type: "opaquehash"},
				{Name: "payload", Type: "bytesequence"},
				{Name: "refine_gas_limit", Type: "gas"},
				{Name: "accumulate_gas_limit", Type: "gas"},
				{Name: "import_segments", Type: "vec<importspec>"},
				{Name: "extrinsic", Type: "vec<extrinsicspec>"},
				{Name: "export_count", Type: "u16"},
			}
			return types.NewStruct(maps)
		},
		"workpackage": func() types.IType {
			maps := []types.TypeMap{
				{Name: "authorization", Type: "bytesequence"},
				{Name: "auth_code_host", Type: "serviceid"},
				{Name: "authorizer", Type: "authorizer"},
				{Name: "context", Type: "refinecontext"},
				{Name: "items", Type: "vec<workitem>"},
			}
			return types.NewStruct(maps)
		},
		"workexecresult": func() types.IType {
			m := map[int]types.TypeMap{
				0: {Name: "ok", Type: "bytesequence"},
				1: {Name: "out_of_gas", Type: "null"},
				2: {Name: "panic", Type: "null"},
				3: {Name: "bad_code", Type: "null"},
				4: {Name: "code_oversize", Type: "null"},
			}
			return types.NewEnum(m)
		},
		"workresult": func() types.IType {
			maps := []types.TypeMap{
				{Name: "service_id", Type: "serviceid"},
				{Name: "code_hash", Type: "opaquehash"},
				{Name: "payload_hash", Type: "opaquehash"},
				{Name: "accumulate_gas", Type: "gas"},
				{Name: "result", Type: "workexecresult"},
			}
			return types.NewStruct(maps)
		},
		"workpackagespec": func() types.IType {
			maps := []types.TypeMap{
				{Name: "hash", Type: "workpackagehash"},
				{Name: "length", Type: "u32"},
				{Name: "erasure_root", Type: "erasureroot"},
				{Name: "exports_root", Type: "exportsroot"},
				{Name: "exports_count", Type: "u16"},
			}
			return types.NewStruct(maps)
		},
		"segmenttreeroot": NewOpaqueHash,
		"segmentrootlookupitem": func() types.IType {
			maps := []types.TypeMap{
				{Name: "work_package_hash", Type: "workpackagehash"},
				{Name: "segment_tree_root", Type: "segmenttreeroot"},
			}
			return types.NewStruct(maps)
		},
		"segmentrootlookups": func() types.IType { return types.NewVec("segmentrootlookupitem") },
		"authorizeroutput":   types.NewHexBytes,
		"workreport": func() types.IType {
			maps := []types.TypeMap{
				{Name: "package_spec", Type: "workpackagespec"},
				{Name: "context", Type: "refinecontext"},
				{Name: "core_index", Type: "coreindex"},
				{Name: "authorizer_hash", Type: "opaquehash"},
				{Name: "auth_output", Type: "authorizeroutput"},
				{Name: "segment_root_lookup", Type: "segmentrootlookups"},
				{Name: "results", Type: "vec<workresult>"},
			}
			return types.NewStruct(maps)
		},

		// Block History
		"mmrpeak": NewOpaqueHash,
		"mmr": func() types.IType {
			maps := []types.TypeMap{
				{Name: "peaks", Type: "vec<option<mmrpeak>>"},
			}
			return types.NewStruct(maps)
		},
		"reportedworkpackage": func() types.IType {
			maps := []types.TypeMap{
				{Name: "hash", Type: "workreporthash"},
				{Name: "exports_root", Type: "exportsroot"},
			}
			return types.NewStruct(maps)
		},
		"blockinfo": func() types.IType {
			maps := []types.TypeMap{
				{Name: "header_hash", Type: "headerhash"},
				{Name: "mmr", Type: "mmr"},
				{Name: "state_root", Type: "stateroot"},
				{Name: "reported", Type: "vec<reportedworkpackage>"},
			}
			return types.NewStruct(maps)
		},

		// Tickets
		"ticketid":      NewOpaqueHash,
		"ticketattempt": types.NewU8,
		"ticketenvelope": func() types.IType {
			maps := []types.TypeMap{
				{Name: "attempt", Type: "ticketattempt"},
				{Name: "signature", Type: "bandersnatchringvrfsignature"},
			}
			return types.NewStruct(maps)
		},
		"ticketbody": func() types.IType {
			maps := []types.TypeMap{
				{Name: "id", Type: "ticketid"},
				{Name: "attempt", Type: "ticketattempt"},
			}
			return types.NewStruct(maps)
		},
		"ticketsbodies": func() types.IType { return types.NewFixedArray(EpochLength, "ticketbody") },
		"epochkeys":     func() types.IType { return types.NewFixedArray(EpochLength, "bandersnatchpublic") },
		"ticketsorkeys": func() types.IType {
			m := map[int]types.TypeMap{
				0: {Name: "tickets", Type: "ticketsbodies"},
				1: {Name: "keys", Type: "epochkeys"},
			}
			return types.NewEnum(m)
		},
		"ticketsextrinsic": func() types.IType { return types.NewVec("ticketenvelope") },

		// Disputes
		"judgement": func() types.IType {
			maps := []types.TypeMap{
				{Name: "vote", Type: "bool"},
				{Name: "index", Type: "validatorindex"},
				{Name: "signature", Type: "ed25519signature"},
			}
			return types.NewStruct(maps)
		},
		"Judgements": func() types.IType { return types.NewFixedArray(ValidatorsSuperMajority, "judgement") },
		"verdict": func() types.IType {
			maps := []types.TypeMap{
				{Name: "target", Type: "workreporthash"},
				{Name: "age", Type: "epochindex"},
				{Name: "votes", Type: "judgements"},
			}
			return types.NewStruct(maps)
		},
		"culprit": func() types.IType {
			maps := []types.TypeMap{
				{Name: "target", Type: "workreporthash"},
				{Name: "key", Type: "ed25519public"},
				{Name: "signature", Type: "ed25519signature"},
			}
			return types.NewStruct(maps)
		},
		"fault": func() types.IType {
			maps := []types.TypeMap{
				{Name: "target", Type: "workreporthash"},
				{Name: "vote", Type: "bool"},
				{Name: "key", Type: "ed25519public"},
				{Name: "signature", Type: "ed25519signature"},
			}
			return types.NewStruct(maps)
		},
		"disputesrecords": func() types.IType {
			maps := []types.TypeMap{
				{Name: "good", Type: "vec<workreporthash>"},
				{Name: "bad", Type: "vec<workreporthash>"},
				{Name: "wonky", Type: "vec<workreporthash>"},
				{Name: "offenders", Type: "vec<ed25519public>"},
			}
			return types.NewStruct(maps)
		},
		"disputesextrinsic": func() types.IType {
			maps := []types.TypeMap{
				{Name: "verdicts", Type: "vec<verdict>"},
				{Name: "culprits", Type: "vec<culprit>"},
				{Name: "faults", Type: "vec<fault>"},
			}
			return types.NewStruct(maps)
		},

		// Preimages
		"preimage": func() types.IType {
			maps := []types.TypeMap{
				{Name: "requester", Type: "serviceid"},
				{Name: "blob", Type: "bytesequence"},
			}
			return types.NewStruct(maps)
		},
		"preimagesextrinsic": func() types.IType { return types.NewVec("preimage") },

		// Assurances
		"bitfield": func() types.IType { return types.NewFixedArray(AvailBitfieldBytes, "u8") },
		"availassurance": func() types.IType {
			maps := []types.TypeMap{
				{Name: "anchor", Type: "opaquehash"},
				{Name: "bitfield", Type: "bitfield"},
				{Name: "validator_index", Type: "validatorindex"},
				{Name: "signature", Type: "ed25519signature"},
			}
			return types.NewStruct(maps)
		},
		"assurancesextrinsic": func() types.IType { return types.NewVec("availassurance") },

		// Guarantees
		"validatorsignature": func() types.IType {
			maps := []types.TypeMap{
				{Name: "validator_index", Type: "validatorindex"},
				{Name: "signature", Type: "ed25519signature"},
			}
			return types.NewStruct(maps)
		},
		"guaranteesignature": func() types.IType {
			maps := []types.TypeMap{
				{Name: "validator_index", Type: "validatorindex"},
				{Name: "signature", Type: "ed25519signature"},
			}
			return types.NewStruct(maps)
		},
		"guaranteesignatures": func() types.IType { return types.NewVec("guaranteesignature") },
		"reportguarantee": func() types.IType {
			maps := []types.TypeMap{
				{Name: "report", Type: "workreport"},
				{Name: "slot", Type: "timeslot"},
				{Name: "signatures", Type: "guaranteesignatures"},
			}
			return types.NewStruct(maps)
		},
		"guaranteesextrinsic": func() types.IType { return types.NewVec("reportguarantee") },

		// Header
		"validators": func() types.IType { return types.NewFixedArray(ValidatorsCount, "bandersnatchpublic") },
		"epochmark": func() types.IType {
			maps := []types.TypeMap{
				{Name: "entropy", Type: "entropy"},
				{Name: "tickets_entropy", Type: "entropy"},
				{Name: "validators", Type: "validators"},
			}
			return types.NewStruct(maps)
		},
		"ticketsmark":   func() types.IType { return types.NewFixedArray(EpochLength, "ticketbody") },
		"offendersmark": func() types.IType { return types.NewVec("ed25519public") },
		"header": func() types.IType {
			maps := []types.TypeMap{
				{Name: "parent", Type: "headerhash"},
				{Name: "parent_state_root", Type: "stateroot"},
				{Name: "extrinsic_hash", Type: "opaquehash"},
				{Name: "slot", Type: "timeslot"},
				{Name: "epoch_mark", Type: "option<epochmark>"},
				{Name: "tickets_mark", Type: "option<ticketsmark>"},
				{Name: "offenders_mark", Type: "offendersmark"},
				{Name: "author_index", Type: "validatorindex"},
				{Name: "entropy_source", Type: "bandersnatchvrfsignature"},
				{Name: "seal", Type: "bandersnatchvrfsignature"},
			}
			return types.NewStruct(maps)
		},

		// Block
		"extrinsic": func() types.IType {
			maps := []types.TypeMap{
				{Name: "tickets", Type: "ticketsextrinsic"},
				{Name: "preimages", Type: "preimagesextrinsic"},
				{Name: "guarantees", Type: "guaranteesextrinsic"},
				{Name: "assurances", Type: "assurancesextrinsic"},
				{Name: "disputes", Type: "disputesextrinsic"},
			}
			return types.NewStruct(maps)
		},
		"block": func() types.IType {
			maps := []types.TypeMap{
				{Name: "header", Type: "header"},
				{Name: "extrinsic", Type: "extrinsic"},
			}
			return types.NewStruct(maps)
		},
	}

	types.RegisterType(m)
}

func NewByteArray(elementCount int) types.IType {
	return types.NewFixedArray(elementCount, "u8")
}

func NewOpaqueHash() types.IType {
	return types.NewFixedArray(32, "u8")
}
