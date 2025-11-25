package builder

import (
	"github.com/ivanzzeth/polymarket-go-order-utils/pkg/model"
)

//go:generate mockery --name ExchangeOrderBuilder
type ExchangeOrderBuilder interface {
	// build an order object including the signature.
	//
	// @param signer - the signer instance to use for signing
	//
	// @param orderData
	//
	// @returns a SignedOrder object (order + signature)
	BuildSignedOrder(signer Signer, orderData *model.OrderData, contract model.VerifyingContract) (*model.SignedOrder, error)

	// Creates an Order object from order data.
	//
	// @param orderData
	//
	// @returns a Order object (not signed)
	BuildOrder(orderData *model.OrderData) (*model.Order, error)

	// Generates the hash of the order from a EIP712TypedData object.
	//
	// @param Order
	//
	// @returns a OrderHash that is a 'common.Hash'
	BuildOrderHash(order *model.Order, contract model.VerifyingContract) (model.OrderHash, error)

	// signs an order
	//
	// @param signer - the signer instance to use for signing
	//
	// @param Order
	//
	// @returns a OrderSignature that is []byte
	BuildOrderSignature(signer Signer, order *model.Order, contract model.VerifyingContract) (model.OrderSignature, error)

	// signs an order
	//
	// @param signer - the signer instance to use for signing
	//
	// @param order hash
	//
	// @returns a OrderSignature that is []byte
	// BuildOrderSignature(signer signer.Signer, orderHash model.OrderHash) (model.OrderSignature, error)
}
