package builder

import (
	"errors"

	"github.com/ivanzzeth/ethsig"
)

var ErrInvalidSignatureLen = errors.New("invalid signature length")

// Signer is an interface for signing hashed data
type Signer interface {
	ethsig.TypedDataSigner
	ethsig.AddressGetter
}
