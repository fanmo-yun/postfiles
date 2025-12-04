package client

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"postfiles/protocol"
	"postfiles/utils"
	"strings"
)

func (c *Client) handleTransfer(conn net.Conn) {
	reader := bufio.NewReaderSize(conn, 32*1024)
	writer := bufio.NewWriterSize(conn, 32*1024)

	if fetchErr := c.fetchList(reader); fetchErr != nil {
		slog.Error("fetch list failed", "err", fetchErr)
		return
	}

	if !c.printList() || !c.askConfirm() {
		slog.Info("user aborted download")
		return
	}

	if sendErr := c.sendConfirm(writer); sendErr != nil {
		slog.Error("send confirm failed", "err", sendErr)
		return
	}

	for len(c.fileMap) > 0 {
		if recvErr := c.recvFile(reader, writer); recvErr != nil {
			if errors.Is(recvErr, os.ErrExist) {
				continue
			}
			slog.Error("file receive failed", "err", recvErr)
			return
		}
	}
}

func (c *Client) fetchList(reader *bufio.Reader) error {
	for {
		recvPkt := new(protocol.Packet)
		if readErr := recvPkt.ReadAndDecode(reader); readErr != nil {
			return readErr
		}

		switch recvPkt.DataType {
		case protocol.FileQuantity:
			c.fileMap[recvPkt.FileName] = recvPkt.FileSize
		case protocol.EndOfTransmission:
			return nil
		default:
			return fmt.Errorf("unknown message type: %d", recvPkt.DataType)
		}
	}
}

func (c *Client) printList() bool {
	if len(c.fileMap) == 0 {
		fmt.Println("No files to download")
		return false
	}

	totalSize := int64(0)
	for name, size := range c.fileMap {
		fmt.Printf("[%-20s] %10s\n", name, utils.ToReadableSize(size))
		totalSize += size
	}
	fmt.Printf("[Total] %d files, %s\n", len(c.fileMap), utils.ToReadableSize(totalSize))

	return true
}

func (c *Client) askConfirm() bool {
	fmt.Print("Confirm accept[Y/n]: ")
	confirm, readinErr := utils.Readin()
	if readinErr != nil {
		return false
	}

	confirm = strings.ToLower(confirm)
	return confirm == "y" || confirm == "yes"
}

func (c *Client) sendConfirm(writer *bufio.Writer) error {
	confirmPkt := protocol.NewPacket(protocol.ConfirmAccept, "", 0)
	if confirmErr := confirmPkt.EncodeAndWrite(writer); confirmErr != nil {
		return confirmErr
	}
	return writer.Flush()
}

func (c *Client) recvFile(reader *bufio.Reader, writer *bufio.Writer) error {
	metaPkt := new(protocol.Packet)
	if metaErr := metaPkt.ReadAndDecode(reader); metaErr != nil {
		return metaErr
	}

	filesize, ok := c.fileMap[metaPkt.FileName]
	if !ok {
		return fmt.Errorf("file not found in client: %s", metaPkt.FileName)
	}

	deleteFlag := true
	defer func() {
		if deleteFlag {
			delete(c.fileMap, metaPkt.FileName)
		}
	}()

	if fstat, err := c.saveDir.Stat(metaPkt.FileName); err == nil && !fstat.IsDir() {
		fmt.Printf("--skip-- %s <-- %s\n", metaPkt.FileName, os.ErrExist)
		rejPkt := protocol.NewPacket(protocol.RejectFile, "", 0)
		if writeErr := rejPkt.EncodeAndWrite(writer); writeErr != nil {
			return writeErr
		}
		if flushErr := writer.Flush(); flushErr != nil {
			return flushErr
		}
		deleteFlag = false
		return os.ErrExist
	}

	accPkt := protocol.NewPacket(protocol.AcceptFile, "", 0)
	if writeErr := accPkt.EncodeAndWrite(writer); writeErr != nil {
		return writeErr
	}
	if flushErr := writer.Flush(); flushErr != nil {
		return flushErr
	}

	file, createErr := c.saveDir.Create(metaPkt.FileName)
	if createErr != nil {
		deleteFlag = false
		return fmt.Errorf("%s cannot be create: %s", metaPkt.FileName, createErr)
	}
	defer file.Close()

	pb, pbErr := utils.NewBar(filesize, metaPkt.FileName)
	if pbErr != nil {
		deleteFlag = false
		rejPkt := protocol.NewPacket(protocol.RejectFile, "", 0)
		if writeErr := rejPkt.EncodeAndWrite(writer); writeErr != nil {
			return writeErr
		}
		if flushErr := writer.Flush(); flushErr != nil {
			return flushErr
		}
		return fmt.Errorf("progress bar error: %w", pbErr)
	}
	mw := io.MultiWriter(file, pb)
	if n, copyErr := io.CopyN(mw, reader, filesize); copyErr != nil {
		deleteFlag = false
		if errors.Is(copyErr, io.EOF) {
			return fmt.Errorf("unexpected EOF: got %d bytes, expected %d", n, filesize)
		}
		return fmt.Errorf("copy file data failed: %w", copyErr)
	}
	return nil
}

func (c *Client) ValidateWritable() error {
	file, tmpErr := os.CreateTemp(c.saveDir.Name(), "write.tmp")
	if tmpErr != nil {
		return fmt.Errorf("%s path is not writable", c.saveDir.Name())
	}
	file.Close()
	if removeErr := os.Remove(file.Name()); removeErr != nil {
		return fmt.Errorf("%s path is not writable", c.saveDir.Name())
	}
	return nil
}
