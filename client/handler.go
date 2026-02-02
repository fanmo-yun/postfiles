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

	if err := c.fetchList(reader); err != nil {
		slog.Error("fetch list failed", "err", err)
		return
	}

	if !c.printList() || !c.askConfirm() {
		slog.Info("user aborted download")
		return
	}

	if err := c.sendConfirm(writer); err != nil {
		slog.Error("send confirm failed", "err", err)
		return
	}

	for len(c.fileMap) > 0 {
		if err := c.recvFile(reader, writer); err != nil {
			if errors.Is(err, os.ErrExist) {
				continue
			}
			slog.Error("file receive failed", "err", err)
			return
		}
	}
}

func (c *Client) fetchList(reader *bufio.Reader) error {
	for {
		recvPkt := new(protocol.Packet)
		if err := recvPkt.ReadAndDecode(reader); err != nil {
			return err
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
	confirm, err := utils.Readin()
	if err != nil {
		return false
	}

	confirm = strings.ToLower(confirm)
	return confirm == "y" || confirm == "yes"
}

func (c *Client) sendConfirm(writer *bufio.Writer) error {
	confirmPkt := protocol.NewPacket(protocol.ConfirmAccept, "", 0)
	if err := confirmPkt.EncodeAndWrite(writer); err != nil {
		return err
	}
	return writer.Flush()
}

func (c *Client) recvFile(reader *bufio.Reader, writer *bufio.Writer) error {
	metaPkt := new(protocol.Packet)
	if err := metaPkt.ReadAndDecode(reader); err != nil {
		return err
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
		if err := rejPkt.EncodeAndWrite(writer); err != nil {
			return err
		}
		if err := writer.Flush(); err != nil {
			return err
		}
		deleteFlag = false
		return os.ErrExist
	}

	accPkt := protocol.NewPacket(protocol.AcceptFile, "", 0)
	if err := accPkt.EncodeAndWrite(writer); err != nil {
		return err
	}
	if err := writer.Flush(); err != nil {
		return err
	}

	file, err := c.saveDir.Create(metaPkt.FileName)
	if err != nil {
		deleteFlag = false
		return fmt.Errorf("%s cannot be create: %s", metaPkt.FileName, err)
	}
	defer file.Close()

	pb, err := utils.NewBar(filesize, metaPkt.FileName)
	if err != nil {
		deleteFlag = false
		rejPkt := protocol.NewPacket(protocol.RejectFile, "", 0)
		if err := rejPkt.EncodeAndWrite(writer); err != nil {
			return err
		}
		if err := writer.Flush(); err != nil {
			return err
		}
		return fmt.Errorf("progress bar error: %w", err)
	}
	mw := io.MultiWriter(file, pb)
	if n, err := io.CopyN(mw, reader, filesize); err != nil {
		deleteFlag = false
		if errors.Is(err, io.EOF) {
			return fmt.Errorf("unexpected EOF: got %d bytes, expected %d", n, filesize)
		}
		return fmt.Errorf("copy file data failed: %w", err)
	}
	return nil
}

func (c *Client) ValidateWritable() error {
	file, err := os.CreateTemp(c.saveDir.Name(), "write.tmp")
	if err != nil {
		return fmt.Errorf("%s path is not writable", c.saveDir.Name())
	}
	file.Close()
	if err := os.Remove(file.Name()); err != nil {
		return fmt.Errorf("%s path is not writable", c.saveDir.Name())
	}
	return nil
}
