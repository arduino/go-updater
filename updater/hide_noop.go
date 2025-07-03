//go:build !windows
// +build !windows

package updater

//nolint:unused
func hideFile(_ string) error {
	return nil
}
