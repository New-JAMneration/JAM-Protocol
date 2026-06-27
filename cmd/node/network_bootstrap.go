package main

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"slices"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/chainspec"
	cehandler "github.com/New-JAMneration/JAM-Protocol/internal/networking/handler/ce"
	uphandler "github.com/New-JAMneration/JAM-Protocol/internal/networking/handler/up"
	"github.com/New-JAMneration/JAM-Protocol/internal/networking/quic"
	validatorpkg "github.com/New-JAMneration/JAM-Protocol/internal/networking/validator"
	nodepkg "github.com/New-JAMneration/JAM-Protocol/internal/node"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

type nodeRole string

const (
	fullNodeRole      nodeRole = "full"
	validatorNodeRole nodeRole = "validator"
)

type nodeRuntime struct {
	peer        *quic.Peer
	syncManager *nodepkg.SyncManager
}

func (n *nodeRuntime) Close() {
	if n.syncManager != nil {
		n.syncManager.Close()
	}
	if n.peer != nil {
		if err := n.peer.Close(); err != nil {
			log.Printf("network peer close failed: %v", err)
		}
	}
}

func startNodeNetworking(ctx context.Context, chainPath, listenAddr, roleFlag string) (*nodeRuntime, error) {
	role, privateKey, err := resolveNodeIdentity(roleFlag)
	if err != nil {
		return nil, err
	}

	udpAddr, err := net.ResolveUDPAddr("udp", listenAddr)
	if err != nil {
		return nil, fmt.Errorf("resolve listen addr: %w", err)
	}

	chain := blockchain.GetInstance()
	eventBus := quic.NewEventBus()
	upHandler := quic.NewDefaultUPHandler()
	ceHandler := quic.NewDefaultCEHandler(chain)

	peer, err := quic.NewPeer(quic.PeerConfig{
		Role:          quic.Validator,
		Addr:          udpAddr,
		GenesisHeader: chain.GenesisBlockHash(),
		PrivateKey:    privateKey,
		UPHandler:     upHandler,
		CEHandler:     ceHandler,
	})
	if err != nil {
		return nil, fmt.Errorf("create networking peer: %w", err)
	}
	peer.SetEventBus(eventBus)

	registerRequiredCEHandlers(peer, chain)
	registerUP0Handler(peer, chain, role, nil)
	if err := peer.Start(ctx); err != nil {
		_ = peer.Close()
		return nil, fmt.Errorf("start networking peer: %w", err)
	}

	syncManager := nodepkg.NewSyncManager(chain, eventBus)
	syncManager.Start()

	if err := bootstrapFromChainSpec(peer, chainPath); err != nil {
		_ = peer.Close()
		syncManager.Close()
		return nil, err
	}

	log.Printf("node role mode: %s", role)
	log.Printf("node networking started at %s", peer.Listener.ListenAddress())
	return &nodeRuntime{
		peer:        peer,
		syncManager: syncManager,
	}, nil
}

func registerRequiredCEHandlers(peer *quic.Peer, chain blockchain.Blockchain) {
	peer.RegisterHandler(byte(cehandler.BlockRequest), func(ctx context.Context, stream *quic.Stream, peerKey ed25519.PublicKey) error {
		return cehandler.HandleBlockRequestStream(chain, stream)
	})
	peer.RegisterHandler(byte(cehandler.StateRequest), func(ctx context.Context, stream *quic.Stream, peerKey ed25519.PublicKey) error {
		return cehandler.HandleStateRequestStream(chain, stream)
	})
}

func registerUP0Handler(peer *quic.Peer, chain *blockchain.ChainState, role nodeRole, vm *validatorpkg.ValidatorManager) {
	up0 := &uphandler.UP0Handler{
		Blocks: func() []types.Block {
			finalized := chain.GetFinalizedBlocks()
			unfinalized := chain.GetUnfinalizedBlocks()
			blocks := make([]types.Block, 0, len(finalized)+len(unfinalized))
			blocks = append(blocks, finalized...)
			blocks = append(blocks, unfinalized...)
			return blocks
		},
		Finalized: func() (types.HeaderHash, error) {
			block := chain.GetLatestFinalizedBlock()
			return hash.ComputeBlockHeaderHash(block.Header)
		},
	}
	peer.RegisterHandler(uphandler.StreamKindUP0, up0.Handle)

	localIsValidator := role == validatorNodeRole
	peer.SetShouldOpenUP0(func(peerKey ed25519.PublicKey) bool {
		if vm == nil {
			return true
		}
		return vm.ShouldOpenUP0(types.Ed25519Public(peerKey), localIsValidator)
	})
}

func bootstrapFromChainSpec(peer *quic.Peer, chainPath string) error {
	if chainPath == "" {
		return nil
	}
	spec, err := blockchain.GetChainSpecFromJson(chainPath)
	if err != nil {
		return fmt.Errorf("load chainspec for bootstrap: %w", err)
	}

	selfKey := peer.Ed25519Key
	for _, raw := range spec.Bootnodes {
		bn, err := chainspec.ParseBootnode(raw)
		if err != nil {
			log.Printf("skip invalid bootnode %q: %v", raw, err)
			continue
		}

		if bytes.Equal(selfKey, bn.PubKey[:]) {
			continue
		}

		addr := &net.UDPAddr{
			IP:   bn.IP,
			Port: int(bn.Port),
		}
		if _, err := peer.Connect(addr, quic.Validator); err != nil {
			log.Printf("bootstrap dial failed (%s): %v", addr.String(), err)
			continue
		}
		log.Printf("bootstrap dial attempted: %s", addr.String())
	}
	return nil
}

type localKeyData struct {
	PrivateKey string `json:"private_key"`
}

var errNoValidatorKey = errors.New("no validator key found")

func resolveNodeIdentity(roleFlag string) (nodeRole, ed25519.PrivateKey, error) {
	switch roleFlag {
	case "":
		return resolveNodeIdentityAuto()
	case string(fullNodeRole):
		return resolveNodeIdentityFull()
	case string(validatorNodeRole):
		return resolveNodeIdentityValidator()
	default:
		return "", nil, fmt.Errorf("invalid --role %q: want full or validator", roleFlag)
	}
}

func resolveNodeIdentityAuto() (nodeRole, ed25519.PrivateKey, error) {
	privateKey, err := loadFirstLocalEd25519Key()
	if err == nil {
		return validatorNodeRole, privateKey, nil
	}
	if !errors.Is(err, errNoValidatorKey) {
		log.Printf("keystore probe failed, fallback to full node key: %v", err)
	}
	return resolveNodeIdentityFull()
}

func resolveNodeIdentityFull() (nodeRole, ed25519.PrivateKey, error) {
	_, generated, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", nil, fmt.Errorf("generate full node key: %w", err)
	}
	return fullNodeRole, generated, nil
}

func resolveNodeIdentityValidator() (nodeRole, ed25519.PrivateKey, error) {
	privateKey, err := loadFirstLocalEd25519Key()
	if err != nil {
		return "", nil, fmt.Errorf("validator role requires keystore/ed25519 key: %w", err)
	}
	return validatorNodeRole, privateKey, nil
}

func loadFirstLocalEd25519Key() (ed25519.PrivateKey, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	keysFile := filepath.Join(cwd, "keystore", "ed25519", "keys.json")
	payload, err := os.ReadFile(keysFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, errNoValidatorKey
		}
		return nil, err
	}

	keys := map[string]localKeyData{}
	if err := json.Unmarshal(payload, &keys); err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		return nil, errNoValidatorKey
	}

	orderedKeys := make([]string, 0, len(keys))
	for pub := range keys {
		orderedKeys = append(orderedKeys, pub)
	}
	slices.Sort(orderedKeys)

	for _, pub := range orderedKeys {
		raw := keys[pub].PrivateKey
		decoded, err := hex.DecodeString(raw)
		if err != nil {
			continue
		}
		switch len(decoded) {
		case ed25519.SeedSize:
			return ed25519.NewKeyFromSeed(decoded), nil
		case ed25519.PrivateKeySize:
			return ed25519.PrivateKey(decoded), nil
		}
	}

	return nil, errNoValidatorKey
}
