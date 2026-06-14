package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveNodeIdentity_Full(t *testing.T) {
	role, key, err := resolveNodeIdentity("full")
	require.NoError(t, err)
	require.Equal(t, fullNodeRole, role)
	require.NotNil(t, key)
}

func TestResolveNodeIdentity_Invalid(t *testing.T) {
	_, _, err := resolveNodeIdentity("auto")
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid --role")
}

func TestResolveNodeIdentity_ValidatorWithoutKeystore(t *testing.T) {
	dir := t.TempDir()
	wd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(wd) })

	_, _, err = resolveNodeIdentity("validator")
	require.Error(t, err)
	require.True(t, errors.Is(err, errNoValidatorKey) || errors.Is(err, os.ErrNotExist))
}

func TestResolveNodeIdentity_ValidatorWithKeystore(t *testing.T) {
	dir := t.TempDir()
	ksDir := filepath.Join(dir, "keystore", "ed25519")
	require.NoError(t, os.MkdirAll(ksDir, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(ksDir, "keys.json"), []byte(`{
  "ab": {"private_key": "0000000000000000000000000000000000000000000000000000000000000001"}
}`), 0o600))

	wd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(wd) })

	role, key, err := resolveNodeIdentity("validator")
	require.NoError(t, err)
	require.Equal(t, validatorNodeRole, role)
	require.NotNil(t, key)
}
