package main

import (
	"os"
	"sort"
	"strings"

	"go.etcd.io/etcd/etcdserver/api/snap"
)

const snapSuffix = ".snap"

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
