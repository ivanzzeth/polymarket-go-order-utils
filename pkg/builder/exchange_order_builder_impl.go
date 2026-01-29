package builder

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ivanzzeth/ethsig"
	"github.com/ivanzzeth/ethsig/eip712"
	"github.com/ivanzzeth/polymarket-go-order-utils/pkg/model"
	"github.com/ivanzzeth/polymarket-go-order-utils/pkg/utils"
)

type ExchangeOrderBuilderImpl struct {
	chainId       *big.Int
	saltGenerator func() int64
}

var _ ExchangeOrderBuilder = (*ExchangeOrderBuilderImpl)(nil)

func NewExchangeOrderBuilderImpl(chainId *big.Int, saltGenerator func() int64) *ExchangeOrderBuilderImpl {
	if saltGenerator == nil {
		saltGenerator = utils.GenerateRandomSalt
	}
	return &ExchangeOrderBuilderImpl{
		chainId:       chainId,
		saltGenerator: saltGenerator,
	}
}

// build an order object including the signature.
//
// @param signer - the signer instance to use for signing
//
// @param orderData
//
// @returns a SignedOrder object (order + signature)
func (e *ExchangeOrderBuilderImpl) BuildSignedOrder(s Signer, orderData *model.OrderData, contract model.VerifyingContract) (*model.SignedOrder, error) {
	order, err := e.BuildOrder(orderData)
	if err != nil {
		return nil, err
	}

	signature, err := e.BuildOrderSignature(s, order, contract)
	if err != nil {
		return nil, err
	}

	// Validate the signature
	orderHash, err := e.BuildOrderHash(order, contract)
	if err != nil {
		return nil, err
	}

	ok, err := ethsig.ValidateSignature(order.Signer, orderHash, signature)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("signature error")
	}

	return &model.SignedOrder{
		Order:     *order,
		Signature: signature,
		OrderHash: orderHash,
	}, nil
}

// Creates an Order object from order data.
//
// @param orderData
//
// @returns a Order object (not signed)
func (e *ExchangeOrderBuilderImpl) BuildOrder(orderData *model.OrderData) (*model.Order, error) {
	var signer common.Address
	if orderData.Signer == "" {
		signer = common.HexToAddress(orderData.Maker)
	} else {
		signer = common.HexToAddress(orderData.Signer)
	}

	var tokenId *big.Int
	var ok bool
	if tokenId, ok = new(big.Int).SetString(orderData.TokenId, 10); !ok {
		return nil, fmt.Errorf("can't parse TokenId: %s as valid *big.Int", orderData.TokenId)
	}

	var makerAmount *big.Int
	if makerAmount, ok = new(big.Int).SetString(orderData.MakerAmount, 10); !ok {
		return nil, fmt.Errorf("can't parse MakerAmount: %s as valid *big.Int", orderData.MakerAmount)
	}

	var takerAmount *big.Int
	if takerAmount, ok = new(big.Int).SetString(orderData.TakerAmount, 10); !ok {
		return nil, fmt.Errorf("can't parse TakerAmount: %s as valid *big.Int", orderData.TakerAmount)
	}

	var expiration *big.Int
	if orderData.Expiration == "" {
		orderData.Expiration = "0"
	}
	if expiration, ok = new(big.Int).SetString(orderData.Expiration, 10); !ok {
		return nil, fmt.Errorf("can't parse Expiration: %s as valid *big.Int", orderData.Expiration)
	}

	var nonce *big.Int
	if nonce, ok = new(big.Int).SetString(orderData.Nonce, 10); !ok {
		return nil, fmt.Errorf("can't parse Nonce: %s as valid *big.Int", orderData.Nonce)
	}

	var feeRateBps *big.Int
	if feeRateBps, ok = new(big.Int).SetString(orderData.FeeRateBps, 10); !ok {
		return nil, fmt.Errorf("can't parse FeeRateBps: %s as valid *big.Int", orderData.FeeRateBps)
	}

	return &model.Order{
		Salt:          new(big.Int).SetInt64(e.saltGenerator()),
		Maker:         common.HexToAddress(orderData.Maker),
		Taker:         common.HexToAddress(orderData.Taker),
		Signer:        signer,
		TokenId:       tokenId,
		MakerAmount:   makerAmount,
		TakerAmount:   takerAmount,
		Side:          new(big.Int).SetInt64(int64(orderData.Side)),
		Expiration:    expiration,
		Nonce:         nonce,
		FeeRateBps:    feeRateBps,
		SignatureType: new(big.Int).SetInt64(int64(orderData.SignatureType)),
	}, nil
}

// Generates the hash of the order from a EIP712TypedData object.
//
// @param Order
//
// @returns a OrderHash that is a 'common.Hash'
func (e *ExchangeOrderBuilderImpl) BuildOrderHash(order *model.Order, contract model.VerifyingContract) (model.OrderHash, error) {
	// Get the verifying contract address
	verifyingContract, err := utils.GetVerifyingContractAddress(e.chainId, contract)
	if err != nil {
		return common.Hash{}, err
	}

	// Build the EIP712 TypedData
	typedData := eip712.TypedData{
		Types: eip712.Types{
			"EIP712Domain": []eip712.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"Order": []eip712.Type{
				{Name: "salt", Type: "uint256"},
				{Name: "maker", Type: "address"},
				{Name: "signer", Type: "address"},
				{Name: "taker", Type: "address"},
				{Name: "tokenId", Type: "uint256"},
				{Name: "makerAmount", Type: "uint256"},
				{Name: "takerAmount", Type: "uint256"},
				{Name: "expiration", Type: "uint256"},
				{Name: "nonce", Type: "uint256"},
				{Name: "feeRateBps", Type: "uint256"},
				{Name: "side", Type: "uint8"},
				{Name: "signatureType", Type: "uint8"},
			},
		},
		PrimaryType: "Order",
		Domain: eip712.TypedDataDomain{
			Name:              "Polymarket CTF Exchange",
			Version:           "1",
			ChainId:           e.chainId.String(),
			VerifyingContract: verifyingContract.Hex(),
		},
		Message: eip712.TypedDataMessage{
			"salt":          order.Salt.String(),
			"maker":         order.Maker.Hex(),
			"signer":        order.Signer.Hex(),
			"taker":         order.Taker.Hex(),
			"tokenId":       order.TokenId.String(),
			"makerAmount":   order.MakerAmount.String(),
			"takerAmount":   order.TakerAmount.String(),
			"expiration":    order.Expiration.String(),
			"nonce":         order.Nonce.String(),
			"feeRateBps":    order.FeeRateBps.String(),
			"side":          fmt.Sprintf("%d", order.Side.Uint64()),
			"signatureType": fmt.Sprintf("%d", order.SignatureType.Uint64()),
		},
	}

	domainSeparator, err := typedData.HashStruct("EIP712Domain", typedData.Domain.Map())
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to hash domain: %w", err)
	}

	typedDataHash, err := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to hash message: %w", err)
	}

	rawData := []byte(fmt.Sprintf("\x19\x01%s%s", string(domainSeparator), string(typedDataHash)))
	orderHash := crypto.Keccak256Hash(rawData)

	return orderHash, nil
}

func (e *ExchangeOrderBuilderImpl) BuildOrderSignature(s Signer, order *model.Order, contract model.VerifyingContract) (model.OrderSignature, error) {
	// Get the verifying contract address
	verifyingContract, err := utils.GetVerifyingContractAddress(e.chainId, contract)
	if err != nil {
		return nil, err
	}

	// Build the EIP712 TypedData
	typedData := eip712.TypedData{
		Types: eip712.Types{
			"EIP712Domain": []eip712.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"Order": []eip712.Type{
				{Name: "salt", Type: "uint256"},
				{Name: "maker", Type: "address"},
				{Name: "signer", Type: "address"},
				{Name: "taker", Type: "address"},
				{Name: "tokenId", Type: "uint256"},
				{Name: "makerAmount", Type: "uint256"},
				{Name: "takerAmount", Type: "uint256"},
				{Name: "expiration", Type: "uint256"},
				{Name: "nonce", Type: "uint256"},
				{Name: "feeRateBps", Type: "uint256"},
				{Name: "side", Type: "uint8"},
				{Name: "signatureType", Type: "uint8"},
			},
		},
		PrimaryType: "Order",
		Domain: eip712.TypedDataDomain{
			Name:              "Polymarket CTF Exchange",
			Version:           "1",
			ChainId:           e.chainId.String(),
			VerifyingContract: verifyingContract.Hex(),
		},
		Message: eip712.TypedDataMessage{
			"salt":          order.Salt.String(),
			"maker":         order.Maker.Hex(),
			"signer":        order.Signer.Hex(),
			"taker":         order.Taker.Hex(),
			"tokenId":       order.TokenId.String(),
			"makerAmount":   order.MakerAmount.String(),
			"takerAmount":   order.TakerAmount.String(),
			"expiration":    order.Expiration.String(),
			"nonce":         order.Nonce.String(),
			"feeRateBps":    order.FeeRateBps.String(),
			"side":          fmt.Sprintf("%d", order.Side.Uint64()),
			"signatureType": fmt.Sprintf("%d", order.SignatureType.Uint64()),
		},
	}

	// Sign the typed data
	signature, err := s.SignTypedData(typedData)
	if err != nil {
		return nil, err
	}

	return signature, nil
}
