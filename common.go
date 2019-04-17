package main

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/rfyiamcool/raft_kvs/consensus/snap"
	"github.com/rfyiamcool/raft_kvs/consensus/utils/fileutil"
	"github.com/rfyiamcool/raft_kvs/consensus/wal"
)

var (
	snapSuffix = ".snap"

	errBadWALName = errors.New("bad wal name")
)

// getSnapNames returns the filename of the snapshots in logical time order (from newest to oldest).
// If there is no available snapshots, an ErrNoSnapshot will be returned.
func getSnapNames(d string) ([]string, error) {
	dir, err := os.Open(d)
	if err != nil {
		return nil, err
	}

	defer dir.Close()
	names, err := dir.Readdirnames(-1)
	if err != nil {
		return nil, err
	}

	snaps := checkSuffix(names)
	if len(snaps) == 0 {
		return nil, snap.ErrNoSnapshot
	}

	sort.Sort(sort.Reverse(sort.StringSlice(snaps)))
	return snaps, nil
}

func checkSuffix(names []string) []string {
	snaps := []string{}
	for i := range names {
		if strings.HasSuffix(names[i], snapSuffix) {
			snaps = append(snaps, names[i])
		}
	}

	return snaps
}

func readWALNames(dirpath string) ([]string, error) {
	names, err := fileutil.ReadDir(dirpath)
	if err != nil {
		return nil, err
	}

	wnames := checkWalNames(names)
	if len(wnames) == 0 {
		return nil, wal.ErrFileNotFound
	}
	return wnames, nil
}

func checkWalNames(names []string) []string {
	wnames := make([]string, 0)
	for _, name := range names {
		_, _, err := parseWALName(name)
		if err != nil {
			continue
		}

		wnames = append(wnames, name)
	}
	return wnames
}

func parseWALName(str string) (seq, index uint64, err error) {
	if !strings.HasSuffix(str, ".wal") {
		return 0, 0, errBadWALName
	}
	_, err = fmt.Sscanf(str, "%016x-%016x.wal", &seq, &index)
	return seq, index, err
}

func walName(seq, index uint64) string {
	return fmt.Sprintf("%016x-%016x.wal", seq, index)
}

func searchIndex(names []string, index uint64) (int, bool) {
	for i := len(names) - 1; i >= 0; i-- {
		name := names[i]
		_, curIndex, err := parseWALName(name)
		if err != nil {
			continue
		}
		if index >= curIndex {
			return i, true
		}
	}

	return -1, false
}
