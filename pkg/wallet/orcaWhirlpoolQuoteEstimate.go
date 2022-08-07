package wallet

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/Dcaf-Protocol/drip-keeper/configs"
)

type QuoteEstimate struct {
	EstimatedAmountIn      string `json:"estimatedAmountIn"`
	EstimatedAmountOut     string `json:"estimatedAmountOut"`
	EstimatedEndTickIndex  int    `json:"estimatedEndTickIndex"`
	EstimatedEndSqrtPrice  string `json:"estimatedEndSqrtPrice"`
	EstimatedFeeAmount     string `json:"estimatedFeeAmount"`
	Amount                 string `json:"amount"`
	AmountSpecifiedIsInput bool   `json:"amountSpecifiedIsInput"`
	AToB                   bool   `json:"aToB"`
	OtherAmountThreshold   string `json:"otherAmountThreshold"`
	SqrtPriceLimit         string `json:"sqrtPriceLimit"`
	TickArray0             string `json:"tickArray0"`
	TickArray1             string `json:"tickArray1"`
	TickArray2             string `json:"tickArray2"`
	Error                  string `json:"error"`
}

func GetOrcaWhirlpoolQuoteEstimate(
	config string,
	tokenAMint string,
	tokenBMint string,
	inputToken string,
	tickSpacing uint16,
	env configs.Environment,
) (QuoteEstimate, error) {
	root := configs.GetProjectRoot()
	scriptPath := fmt.Sprintf("%s/pkg/wallet/orcaWhirlpoolQuoteEstimate.ts", root)
	command := fmt.Sprintf("ts-node %s", scriptPath) +
		fmt.Sprintf(" %s", config) +
		fmt.Sprintf(" %s", tokenAMint) +
		fmt.Sprintf(" %s", tokenBMint) +
		fmt.Sprintf(" %s", inputToken) +
		fmt.Sprintf(" %d", tickSpacing) +
		fmt.Sprintf(" %s", env)
	parts := strings.Fields(command)
	data, err := exec.Command(parts[0], parts[1:]...).Output()
	if err != nil {
		return QuoteEstimate{}, err
	}
	var quote QuoteEstimate
	if err := json.Unmarshal(data, &quote); err != nil {
		return QuoteEstimate{}, fmt.Errorf("failed to unmarshal quote estimate %w", err)
	}
	if quote.Error != "" {
		return QuoteEstimate{}, fmt.Errorf("%s", quote.Error)
	}
	return quote, nil
}
