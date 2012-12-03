// +build !laptop

package main

func laptop()                    {}
func volumeIcon(x string) string { return x }
func getVolume() uint8           { return 0 }
