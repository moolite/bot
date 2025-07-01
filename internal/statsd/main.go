package statsd

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/url"
	"reflect"
	"strconv"
	"sync"
	"time"
)

type BufferSize int

// MTU - 60 ipv6 - 20 header
const (
	BUFFER_SIZE       BufferSize = 1500 - 60 - 20
	BUFFER_SIZE_JUMBO            = 9000 - 60 - 20
	BUFFER_SIZE_SAFE             = 576 - 60 - 20
)

type bufferPool struct {
	pool sync.Pool
}

func newBufferPool(size BufferSize) *bufferPool {
	p := &bufferPool{
		pool: sync.Pool{},
	}

	p.pool.New = func() interface{} {
		b := make([]byte, 0, size)
		return bytes.NewBuffer(b)
	}

	return p
}

func (b *bufferPool) Get() *bytes.Buffer {
	return b.pool.Get().(*bytes.Buffer)
}

func (b *bufferPool) Put(buf *bytes.Buffer) {
	b.pool.Put(buf)
}

type Tag struct {
	Key   string
	Value string
}

func (t *Tag) String() string {
	return fmt.Sprintf("%s:%s", t.Key, t.Value)
}

func (t *Tag) Len() int {
	return len(t.Key) + len(t.Value) + 1
}

type Client struct {
	Prefix string

	tickDuration time.Duration
	conn         net.Conn
	bufferPool   *bufferPool
	m            sync.RWMutex
	statsbuf     *bytes.Buffer
	ErrorHandler func(err error)
	errs         chan (error)
}

func New(url url.URL, connectionTimeout time.Duration, bufferSize BufferSize) (*Client, error) {
	conn, err := net.DialTimeout("udp", url.String(), connectionTimeout)
	if err != nil {
		return nil, err
	}

	c := &Client{conn: conn}
	c.ErrorHandler = NoopErrorHandler
	c.SetbufferSize(bufferSize)

	return c, err
}

func (c *Client) SetbufferSize(bufferSize BufferSize) *Client {
	c.bufferPool = newBufferPool(bufferSize)
	c.errs = make(chan error)
	return c
}

func (c *Client) SetErrorHandler(fun func(error)) *Client {
	c.ErrorHandler = fun
	return c
}

func (c *Client) SetTick(duration time.Duration) *Client {
	c.tickDuration = duration
	return c
}

func (c *Client) Run(ctx context.Context) {
	tick := time.NewTicker(c.tickDuration)
	for {
		select {
		case err := <-c.errs:
			c.ErrorHandler(err)
		case <-tick.C:
			c.Flush()
		case <-ctx.Done():
			c.Flush()
			c.conn.Close()
			return
		default:
			continue
		}
	}
}

// Dogstag
// metric.name:0|c|#tagName:val,tag2Name:val2
func (c *Client) Flush() {
	c.m.Lock()
	buf := c.statsbuf
	c.statsbuf = c.bufferPool.Get()
	c.m.Unlock()

	_, err := buf.WriteTo(c.conn)
	if err != nil {
		c.errs <- err
	}

	buf.Reset()
	c.bufferPool.Put(buf)
}

func (c *Client) Op(key string, value interface{}, op string, tags []Tag) {
	// swap and put back the buffer
	buf := c.bufferPool.Get()

	buf.WriteString(key)
	writeNumberToBuffer(buf, value)
	buf.WriteByte('|')
	buf.WriteString(op)

	if len(tags) > 0 {
		buf.WriteByte('|')
		buf.WriteByte('#')
		for _, t := range tags {
			buf.WriteByte('|')
			buf.WriteString(t.Key)
			buf.WriteByte(':')
			buf.WriteString(t.Value)
		}
	}

	// check the capacity of the underlying buffer including a newline
	if c.statsbuf.Cap()-c.statsbuf.Len()-buf.Len()-1 < 0 {
		c.Flush()
	}

	if c.statsbuf.Len() > 0 {
		c.statsbuf.WriteByte('\n')
	}

	_, err := buf.WriteTo(c.statsbuf)
	if err != nil {
		c.errs <- err
	}

	buf.Reset()
	c.bufferPool.Put(buf)
}

const (
	op_inc = `c`
)

func (c *Client) Inc(key string, value int, tags ...Tag) {
	c.Op(key, strconv.Itoa(value), op_inc, tags)
}

func (c *Client) Gauge(key string, value interface{}, tags ...Tag) {
	c.Op(key, value, "g", tags)
}

func (c *Client) KeyVal(key string, value string, tags ...Tag) {
	c.Op(key, value, "kv", tags)
}

func (c *Client) Set(key string, value string, tags ...Tag) {
	c.Op(key, value, "s", tags)
}

func (c *Client) Counter(key string, value interface{}, tags ...Tag) {
	c.Op(key, value, "s", tags)
}

func (c *Client) Timing(key string, tags ...Tag) func() {
	t := time.Now()
	return func() {
		c.Counter(key, time.Since(t).Nanoseconds(), tags...)
	}
}

func writeNumberToBuffer(buf *bytes.Buffer, value interface{}) error {
	var err error
	switch v := value.(type) {
	case []byte:
		_, err = buf.Write(v)
	case string:
		_, err = fmt.Fprint(buf, v)
	case
		int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		uintptr:
		_, err = fmt.Fprintf(buf, "%d", v)
	case float32, float64:
		_, err = fmt.Fprintf(buf, "%f", v)
	default:
		// Try to handle values that are numeric but not directly type-matched
		rv := reflect.ValueOf(value)
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			_, err = fmt.Fprintf(buf, "%d", rv.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			_, err = fmt.Fprintf(buf, "%d", rv.Uint())
		case reflect.Float32, reflect.Float64:
			_, err = fmt.Fprintf(buf, "%f", rv.Float())
		default:
			return fmt.Errorf("unsupported type: %T", value)
		}
	}

	return err
}

func NoopErrorHandler(_ error) {}
