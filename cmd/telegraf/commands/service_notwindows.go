//go:build !windows

package commands

func getServiceActions() serviceActions { return serviceActions{} }
