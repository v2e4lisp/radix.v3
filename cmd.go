package radix

// Cmd implements the Action interface and describes a single redis command to
// be performed. The Cmd field is the name of the redis command to be performend
// and is always required. Keys are the keys being operated on, and may be left
// empty if the command doesn't operate on any specific key(s) (e.g. SCAN). Args
// are any extra arguments to the command and can be almost any time (TODO flesh
// that statement out).
//
// See the Decoder docs for more on how results are unmarshalled into Rcv.
//
// A Cmd can be filled in manually, or the shortcut methods may be used for more
// convenience. For example to set a key:
//
//	cmd := radix.Cmd{}.C("SET").K("fooKey").A("will be set to this string")
//
// And then to retrieve that key:
//
//	var fooVal string
//	cmd := radix.Cmd{}.C("GET").K("fooKey").R(&fooVal)
type Cmd struct {
	Cmd  []byte
	Keys [][]byte
	Args []interface{}
	Rcv  interface{}
}

// C (short for "Cmd") sets the Cmd field to the given string and returns the
// new Cmd object
func (c Cmd) C(cmd string) Cmd {
	c.Cmd = append(c.Cmd[:0], cmd...)
	return c
}

// K (short for "Key") appends the given key to the Keys field and returns the
// new Cmd object
func (c Cmd) K(key string) Cmd {
	if len(c.Keys) == cap(c.Keys) {
		// if keys is full we'll just have to bite the bullet and allocate the
		// memory
		c.Keys = append(c.Keys, []byte(key))
		return c
	}
	i := len(c.Keys)
	c.Keys = c.Keys[:i+1]
	c.Keys[i] = append(c.Keys[i][:0], key...)
	return c
}

// Key implements the Key method for that Action interface.
func (c Cmd) Key() []byte {
	if len(c.Keys) == 0 {
		return nil
	}
	return c.Keys[0]
}

// A (short for "Arg") appends the given argument to the Args slice and returns
// the new Cmd object.
func (c Cmd) A(arg interface{}) Cmd {
	c.Args = append(c.Args, arg)
	return c
}

// R (short for "Rcv") sets the Rcv field to the given receiver (which should be
// a pointer) and returns the new Cmd object. The receiver is what the response
// to the Cmd is unmarshalled into.
//
// If Rcv isn't set then the command's return value will be discarded.
func (c Cmd) R(v interface{}) Cmd {
	c.Rcv = v
	return c
}

// Reset can be used to reuse a Cmd object. In cases where memory allocations
// are a concern this allows the user of radix to conserve them easily, either
// by always using the same Cmd object within a go-routine or by creating a Cmd
// pool.
//
// The returned Cmd object can be used as if it was just instantiated.
func (c Cmd) Reset() Cmd {
	c.Cmd = c.Cmd[:0]
	c.Keys = c.Keys[:0]
	c.Args = c.Args[:0]
	c.Rcv = nil
	return c
}

// Run implements the Run method of the Action interface. It writes the Cmd to
// the Conn, and unmarshals the result into the Rcv field (if set). It calls
// Close on the Conn if any errors occur.
func (c Cmd) Run(conn Conn) error {
	if err := conn.Encode(c); err != nil {
		conn.Close()
		return err
	} else if err := conn.Decode(c.Rcv); err != nil {
		conn.Close()
		return err
	}
	return nil
}