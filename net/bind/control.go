package bind

import "syscall"

func reuseAddrControl(_, _ string, conn syscall.RawConn) error {
	var e2 error

	e1 := conn.Control(func(fd uintptr) {
		e2 = controlSetReuseAddr(fd)
	})

	if e1 != nil {
		return e1
	}
	return e2
}
