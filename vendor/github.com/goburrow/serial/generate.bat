go tool cgo -godefs types_windows.go | gofmt > ztypes_windows.go
go generate syscall_windows.go
