package client

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"syscall"
	"time"
)

var name string
var xtrn string
var from string
var drop string

const defaultBufferSize = 4096

// TelnetClient represents a TCP client which is responsible for writing input data and printing response.
type TelnetClient struct {
	destination     *net.TCPAddr
	responseTimeout time.Duration
}

// NewTelnetClient method creates new instance of TCP client.
func NewTelnetClient(options Options) *TelnetClient {
	tcpAddr := createTCPAddr(options)
	resolved := resolveTCPAddr(tcpAddr)

	return &TelnetClient{
		destination:     resolved,
		responseTimeout: options.Timeout(),
	}
}

func createTCPAddr(options Options) string {
	dropFileData()

	var buffer bytes.Buffer
	buffer.WriteString(options.Host())
	buffer.WriteByte(':')
	buffer.WriteString(fmt.Sprintf("%d", options.Port()))
	return buffer.String()
}

func resolveTCPAddr(addr string) *net.TCPAddr {
	resolved, error := net.ResolveTCPAddr("tcp", addr)
	if nil != error {
		log.Fatalf("Error occured while resolving TCP address \"%v\": %v\n", addr, error)
	}

	return resolved
}

// ProcessData method processes data: reads from input and writes to output.
func (t *TelnetClient) ProcessData(inputData io.Reader, outputData io.Writer) {
	connection, error := net.DialTCP("tcp", nil, t.destination)
	if nil != error {
		log.Fatalf("Error occured while connecting to address \"%v\": %v\n", t.destination.String(), error)
	}

	defer connection.Close()

	requestDataChannel := make(chan []byte)
	doneChannel := make(chan bool)
	responseDataChannel := make(chan []byte)

	go t.readInputData(inputData, requestDataChannel, doneChannel)
	go t.readServerData(connection, responseDataChannel)

	var afterEOFResponseTicker = new(time.Ticker)
	var afterEOFMode bool
	var somethingRead bool

	for {
		select {
		case request := <-requestDataChannel:
			if _, error := connection.Write(request); nil != error {
				log.Fatalf("Error occured while writing to TCP socket: %v\n", error)
			}

		case <-doneChannel:
			afterEOFMode = true
			afterEOFResponseTicker = time.NewTicker(t.responseTimeout)

		case response := <-responseDataChannel:
			outputData.Write([]byte(fmt.Sprintf("%v", string(response))))
			somethingRead = true

			if afterEOFMode {
				afterEOFResponseTicker.Stop()
				afterEOFResponseTicker = time.NewTicker(t.responseTimeout)
			}

		case <-afterEOFResponseTicker.C:
			if !somethingRead {
				log.Println("Nothing read. Maybe connection timeout.")
			}
			return
		}

	}

}

func (t *TelnetClient) readInputData(inputData io.Reader, toSent chan<- []byte, doneChannel chan<- bool) {
	buffer := make([]byte, defaultBufferSize)
	var error error
	var n int

	reader := bufio.NewReader(inputData)

	for nil == error {
		n, error = reader.Read(buffer)
		toSent <- buffer[:n]
	}

	t.assertEOF(error)
	doneChannel <- true
}

func (t *TelnetClient) readServerData(connection *net.TCPConn, received chan<- []byte) {
	buffer := make([]byte, defaultBufferSize)
	var error error
	var n int

	dropFileData()

	// Send data to Rlogin
	fmt.Fprintf(connection, "\x00%s\x00%s\x00%s\x00", "", name+"-"+from, xtrn)

	for nil == error {
		n, error = connection.Read(buffer)
		received <- buffer[:n]
	}

	go startPolling1(connection)

	t.assertEOF(error)
}

func (t *TelnetClient) assertEOF(error error) {
	if "EOF" != error.Error() {
		log.Fatalf("Error occured while operating on TCP socket: %v\n", error)
	}
}

func dropFileData() {
	from = os.Args[5]
	xtrn = os.Args[4]
	drop := os.Args[3]

	file, err := os.Open(drop)
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var text []string
	for scanner.Scan() {
		text = append(text, scanner.Text())
	}
	file.Close()

	count := 0
	for _, eachLn := range text {
		if count == 35 {
			name = eachLn
		}
		if count == 51 {
			break
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return
}

func connCheck(conn net.Conn) error {
	var sysErr error = nil
	rc, err := conn.(syscall.Conn).SyscallConn()
	if err != nil {
		conn.Close()
		os.Exit(1)
		return err

	}
	err = rc.Read(func(fd uintptr) bool {
		var buf []byte = []byte{0}
		n, _, err := syscall.Recvfrom(int(fd), buf, syscall.MSG_PEEK|syscall.MSG_DONTWAIT)
		switch {
		case n == 0 && err == nil:
			sysErr = io.EOF
			log.Println(sysErr)
			conn.Close()
			os.Exit(1)
		case err == syscall.EAGAIN || err == syscall.EWOULDBLOCK:
			sysErr = nil
			conn.Close()
			os.Exit(1)
			log.Println(sysErr)
		default:
			sysErr = err
			conn.Close()
			os.Exit(1)
			log.Println(err)
		}
		return true
	})
	if err != nil {
		log.Println(err)
		conn.Close()
		os.Exit(1)
		return err
	}
	conn.Close()
	os.Exit(1)
	return sysErr
}

func startPolling1(conn net.Conn) {
	for {
		time.Sleep(500 * time.Millisecond)
		go connCheck(conn)
	}
}
