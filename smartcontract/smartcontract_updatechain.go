package smartcontract

import "github.com/bcbchain/sdk/sdk/types"

func (s *SmartContract) addContractAllKey() {
	if s.sdk.Block().ChainID() == "bcb" {
		allContractAddr := []types.Address{
			"bcbCEstN7YrDYC8Augy4GTMVgXsHeHEprYSE", // token-basic-foundation 1.0
			"bcb7yKaXZZafv7j34xLx8tn3JRQrVZURpYPe", // token-basic-team 1.0
			"bcb4DgJDQzEgv4PvgD2spWASjkYUziBzM4nb", // transferAgency 1.0
			"bcbAWqW2V8kUvMCRmepJqf8A4bWCcBsmDoxw", // token-byb 1.0
			"bcbKA7Lfg9C8hxCdMkLPxugMjZGaX6BF7E5e", // yuebao-dc 1.0
			"bcbFpys9Sy6QRc1p27TgSSAcx1L6cW3nLtXf", // yuebao-dc 2.0
			"bcbDzodYao8d8ZVn9zEhxgC17j7SgYDDqxSV", // yuebao-usdy 1.0
		}
		s._setAllContractAddr(allContractAddr)

	} else if s.sdk.Block().ChainID() == "bcbtest" {
		allContractAddr := []types.Address{
			"bcbtestNVwYSosaBAJu4z373hcxfAKT52AEu3oZZ", // token-basic-foundation 1.0
			"bcbtest6Y5amaZFKf8vQHTcviR1Y5HFW1Wtq4PJh", // token-basic-team 1.0
			"bcbtest57Hxh9KLKi2QU4RztWNAvJvN1GMH38LMR", // yuebao-dc 1.0
			"bcbtestPuggCRTeyTvzv9SwmdWyKQ7VJ199oRvbX", // yuebao-dc 2.0
			"bcbtestGBrJMpSyUyRHkq2qECz6ZoVBhmZ71JwfK", // yuebao-usdy 1.0
		}
		s._setAllContractAddr(allContractAddr)

	} else {
		// do nothing
	}
}
