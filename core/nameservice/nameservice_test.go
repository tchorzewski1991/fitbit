package nameservice_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tchorzewski1991/fitbit/core/nameservice"
)

func TestNameService(t *testing.T) {

	// Ensure name service checks whether accounts path exists.
	_, err := nameservice.New("testdata/accounts/broken")
	assert.EqualError(t, err, "accounts path is not valid")

	// Ensure name service checks whether account file has supported extension.
	_, err = nameservice.New("testdata/accounts/invalid/test.invalid")
	assert.EqualError(t, err, "detected not supported file ext")

	// Ensure name service does not complain when accounts path exists.
	ns, err := nameservice.New("testdata/accounts/valid/")
	assert.Nil(t, err)

	// Ensure name service fallbacks to the given account ID if account name not found.
	name := ns.FindName("not-found")
	assert.Equal(t, "not-found", name)

	// Ensure name service returns account name for existing account ID.
	name = ns.FindName("0x0ee5ba68586c85880B0900D0dEe0eEcBB37010e0")
	assert.Equal(t, "test", name)
}
