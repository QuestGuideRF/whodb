package memcached
import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)
type Client struct {
	conn   net.Conn
	reader *bufio.Reader
}
type Item struct {
	Key        string
	Value      []byte
	Flags      uint32
	Expiration int32
	CAS        uint64
	Size       int
}
type MetadumpEntry struct {
	Key        string
	Expiration int64
	LastAccess int64
	CAS        uint64
	Size       int
	Class      int
}
const dialTimeout = 5 * time.Second
func Dial(address string) (*Client, error) {
	conn, err := net.DialTimeout("tcp", address, dialTimeout)
	if err != nil {
		return nil, fmt.Errorf("memcached dial: %w", err)
	}
	return &Client{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}, nil
}
func DialTLS(address string, tlsConfig *tls.Config) (*Client, error) {
	dialer := &net.Dialer{Timeout: dialTimeout}
	conn, err := tls.DialWithDialer(dialer, "tcp", address, tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("memcached TLS dial: %w", err)
	}
	return &Client{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}, nil
}
func (c *Client) Close() error {
	return c.conn.Close()
}
func (c *Client) Authenticate(username, password string) error {
	payload := username + " " + password
	cmd := fmt.Sprintf("set auth 0 0 %d\r\n%s\r\n", len(payload), payload)
	if _, err := fmt.Fprint(c.conn, cmd); err != nil {
		return fmt.Errorf("memcached auth send: %w", err)
	}
	line, err := c.readLine()
	if err != nil {
		return fmt.Errorf("memcached auth read: %w", err)
	}
	if line != "STORED" {
		return fmt.Errorf("memcached auth failed: %s", line)
	}
	return nil
}
func (c *Client) Version() (string, error) {
	if _, err := fmt.Fprint(c.conn, "version\r\n"); err != nil {
		return "", fmt.Errorf("memcached version send: %w", err)
	}
	line, err := c.readLine()
	if err != nil {
		return "", fmt.Errorf("memcached version read: %w", err)
	}
	if !strings.HasPrefix(line, "VERSION ") {
		return "", fmt.Errorf("memcached version unexpected response: %s", line)
	}
	return strings.TrimPrefix(line, "VERSION "), nil
}
func (c *Client) Ping() error {
	_, err := c.Version()
	return err
}
func (c *Client) Get(key string) (*Item, error) {
	if _, err := fmt.Fprintf(c.conn, "get %s\r\n", key); err != nil {
		return nil, fmt.Errorf("memcached get send: %w", err)
	}
	return c.readGetResponse(false)
}
func (c *Client) Gets(key string) (*Item, error) {
	if _, err := fmt.Fprintf(c.conn, "gets %s\r\n", key); err != nil {
		return nil, fmt.Errorf("memcached gets send: %w", err)
	}
	return c.readGetResponse(true)
}
func (c *Client) Set(item *Item) error {
	return c.storageCommand("set", item)
}
func (c *Client) Add(item *Item) error {
	return c.storageCommand("add", item)
}
func (c *Client) Replace(item *Item) error {
	return c.storageCommand("replace", item)
}
func (c *Client) Delete(key string) error {
	if _, err := fmt.Fprintf(c.conn, "delete %s\r\n", key); err != nil {
		return fmt.Errorf("memcached delete send: %w", err)
	}
	line, err := c.readLine()
	if err != nil {
		return fmt.Errorf("memcached delete read: %w", err)
	}
	if line == "NOT_FOUND" {
		return fmt.Errorf("memcached delete: key not found")
	}
	if line != "DELETED" {
		return fmt.Errorf("memcached delete: %s", line)
	}
	return nil
}
func (c *Client) Stats() (map[string]string, error) {
	if _, err := fmt.Fprint(c.conn, "stats\r\n"); err != nil {
		return nil, fmt.Errorf("memcached stats send: %w", err)
	}
	return c.readStats()
}
func (c *Client) Metadump() ([]MetadumpEntry, error) {
	if _, err := fmt.Fprint(c.conn, "lru_crawler metadump all\r\n"); err != nil {
		return nil, fmt.Errorf("memcached metadump send: %w", err)
	}
	var entries []MetadumpEntry
	for {
		line, err := c.readLine()
		if err != nil {
			return nil, fmt.Errorf("memcached metadump read: %w", err)
		}
		if line == "END" {
			break
		}
		if strings.HasPrefix(line, "CLIENT_ERROR") || strings.HasPrefix(line, "ERROR") {
			return nil, fmt.Errorf("memcached metadump: %s", line)
		}
		entry, err := parseMetadumpLine(line)
		if err != nil {
			continue
		}
		entries = append(entries, entry)
	}
	return entries, nil
}
func (c *Client) storageCommand(cmd string, item *Item) error {
	header := fmt.Sprintf("%s %s %d %d %d\r\n", cmd, item.Key, item.Flags, item.Expiration, len(item.Value))
	if _, err := fmt.Fprint(c.conn, header); err != nil {
		return fmt.Errorf("memcached %s send header: %w", cmd, err)
	}
	if _, err := c.conn.Write(item.Value); err != nil {
		return fmt.Errorf("memcached %s send data: %w", cmd, err)
	}
	if _, err := fmt.Fprint(c.conn, "\r\n"); err != nil {
		return fmt.Errorf("memcached %s send terminator: %w", cmd, err)
	}
	line, err := c.readLine()
	if err != nil {
		return fmt.Errorf("memcached %s read: %w", cmd, err)
	}
	if line != "STORED" {
		return fmt.Errorf("memcached %s: %s", cmd, line)
	}
	return nil
}
func (c *Client) readGetResponse(withCAS bool) (*Item, error) {
	line, err := c.readLine()
	if err != nil {
		return nil, fmt.Errorf("memcached get read header: %w", err)
	}
	if line == "END" {
		return nil, nil
	}
	if !strings.HasPrefix(line, "VALUE ") {
		return nil, fmt.Errorf("memcached get unexpected: %s", line)
	}
	parts := strings.Fields(line)
	expectedParts := 4
	if withCAS {
		expectedParts = 5
	}
	if len(parts) < expectedParts {
		return nil, fmt.Errorf("memcached get malformed VALUE line: %s", line)
	}
	flags, err := strconv.ParseUint(parts[2], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("memcached get parse flags: %w", err)
	}
	byteCount, err := strconv.Atoi(parts[3])
	if err != nil {
		return nil, fmt.Errorf("memcached get parse bytes: %w", err)
	}
	var cas uint64
	if withCAS && len(parts) >= 5 {
		cas, err = strconv.ParseUint(parts[4], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("memcached get parse CAS: %w", err)
		}
	}
	data := make([]byte, byteCount+2)
	if _, err := io.ReadFull(c.reader, data); err != nil {
		return nil, fmt.Errorf("memcached get read data: %w", err)
	}
	endLine, err := c.readLine()
	if err != nil {
		return nil, fmt.Errorf("memcached get read END: %w", err)
	}
	if endLine != "END" {
		return nil, fmt.Errorf("memcached get expected END, got: %s", endLine)
	}
	return &Item{
		Key:   parts[1],
		Value: data[:byteCount],
		Flags: uint32(flags),
		CAS:   cas,
		Size:  byteCount,
	}, nil
}
func (c *Client) readStats() (map[string]string, error) {
	stats := make(map[string]string)
	for {
		line, err := c.readLine()
		if err != nil {
			return nil, fmt.Errorf("memcached stats read: %w", err)
		}
		if line == "END" {
			break
		}
		parts := strings.SplitN(line, " ", 3)
		if len(parts) == 3 && parts[0] == "STAT" {
			stats[parts[1]] = parts[2]
		}
	}
	return stats, nil
}
func (c *Client) readLine() (string, error) {
	line, err := c.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}
func parseMetadumpLine(line string) (MetadumpEntry, error) {
	var entry MetadumpEntry
	fields := strings.Fields(line)
	for _, field := range fields {
		kv := strings.SplitN(field, "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "key":
			decoded, err := url.QueryUnescape(kv[1])
			if err != nil {
				entry.Key = kv[1]
			} else {
				entry.Key = decoded
			}
		case "exp":
			entry.Expiration, _ = strconv.ParseInt(kv[1], 10, 64)
		case "la":
			entry.LastAccess, _ = strconv.ParseInt(kv[1], 10, 64)
		case "cas":
			entry.CAS, _ = strconv.ParseUint(kv[1], 10, 64)
		case "size":
			entry.Size, _ = strconv.Atoi(kv[1])
		case "cls":
			entry.Class, _ = strconv.Atoi(kv[1])
		}
	}
	if entry.Key == "" {
		return entry, fmt.Errorf("metadump line missing key: %s", line)
	}
	return entry, nil
}