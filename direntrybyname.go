package main

import "io/fs"

type dirEntryByName []fs.DirEntry

func (s dirEntryByName) Len() int {
	return len(s)
}
func (s dirEntryByName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s dirEntryByName) Less(i, j int) bool {
	return s[i].Name() < s[j].Name()
}
