package common

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"strconv"
)

func readBytes(buffSize int, conn *tls.Conn) ([]byte, error) {
	buf := make([]byte, buffSize)
	var buff *bytes.Buffer = bytes.NewBuffer([]byte{})
	var n int = 1
	var err error
	n, err = conn.Read(buf)
	if err != nil {
		return nil, err
	}
	if n > 0 {
		_, err = buff.Write(buf[:n])
	}
	return buff.Bytes(), nil
}

func readString(buffSize int, conn *tls.Conn) (string, error) {
	buf := make([]byte, buffSize)
	var buff *bytes.Buffer = bytes.NewBuffer([]byte{})
	var n int = 1
	var err error
	n, err = conn.Read(buf)
	if err != nil {
		return "", err
	}
	if n > 0 {
		_, err = buff.Write(buf[:n])
	}
	return buff.String(), nil
}

func ReadString(conn *tls.Conn) (string, error) {
	return readString(2048, conn)
}

func ReadStringBuffer(buffSize int, conn *tls.Conn) (string, error) {
	return readString(buffSize, conn)
}

func Read(conn *tls.Conn) ([]byte, error) {
	value, errX := readString(2048, conn)
	if errX != nil {
		return nil, errX
	}
	size, errY := strconv.Atoi(value)
	if errY != nil {
		return nil, errY
	}
	return readBytes(size, conn)
}

func writeString(value string, conn *tls.Conn) (int, error) {
	return conn.Write([]byte(value))

}

func writeBytes(value []byte, conn *tls.Conn) (int, error) {
	return conn.Write(value)

}

func WriteString(value string, conn *tls.Conn) (int, error) {
	return writeString(value, conn)
}

func Write(value []byte, conn *tls.Conn) (int, error) {
	n, err := writeString(fmt.Sprintf("%v", len(value)), conn)
	if err != nil {
		return 0, err
	}
	if n == 0 {
		return 0, errors.New(fmt.Sprintf("No bytes written as frame-size"))
	}
	return writeBytes(value, conn)

}
