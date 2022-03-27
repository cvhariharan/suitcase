package utils

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/sys/unix"
)

func Memfile(name string) (int, error) {
	fd, err := unix.MemfdCreate(name, 0)
	if err != nil {
		log.Println(err)
		return -1, err
	}

	err = unix.Ftruncate(fd, 0)
	if err != nil {
		log.Println(err)
		return -1, err
	}

	return fd, nil
}

func FdtoFile(fd int) *os.File {
	return os.NewFile(uintptr(fd), fmt.Sprintf("/proc/self/fd/%d", fd))
}
