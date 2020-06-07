package log_test

import (
	"github.com/stretchr/testify/require"
	api "github.com/tPhume/proglog/api/v1"
	"github.com/tPhume/proglog/internal/log"
	"io/ioutil"
	"os"
	"testing"
)

func TestLog(t *testing.T) {
	for scenario, fn := range map[string]func(
		t *testing.T,
		log *log.Log,
	){
		"append and read a record succeeds": testAppendRead,
		"offset out of range error":         testOutOfRangeErr,
		"init with existing segments":       testInitExisting,
	} {
		t.Run(scenario, func(t *testing.T) {
			dir, err := ioutil.TempDir("", "store-test")
			require.NoError(t, err)
			defer os.RemoveAll(dir)

			c := log.Config{}
			c.Segment.MaxStoreBytes = 32

			l, err := log.NewLog(dir, c)
			require.NoError(t, err)

			fn(t, l)
		})
	}
}

func testAppendRead(t *testing.T, log *log.Log) {
	a := &api.Record{Value: []byte("hello world")}

	off, err := log.Append(a)
	require.NoError(t, err)
	require.Equal(t, uint64(0), off)

	read, err := log.Read(off)
	require.NoError(t, err)
	require.Equal(t, a, read)
}

func testOutOfRangeErr(t *testing.T, log *log.Log) {
	read, err := log.Read(1)
	require.Nil(t, read)
	require.Error(t, err)
}

func testInitExisting(t *testing.T, o *log.Log) {
	a := &api.Record{Value: []byte("hello world")}

	for i := 0; i < 3; i++ {
		_, err := o.Append(a)
		require.NoError(t, err)
	}
	require.NoError(t, o.Close())

	n, err := log.NewLog(o.Dir, o.Config)
	require.NoError(t, err)

	off, err := n.Append(a)
	require.NoError(t, err)
	require.Equal(t, uint64(3), off)
}
