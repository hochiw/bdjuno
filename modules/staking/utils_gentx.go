package staking

import (
	"fmt"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/forbole/bdjuno/v3/modules/staking/keybase"
	"github.com/forbole/bdjuno/v3/modules/utils"
	"github.com/forbole/bdjuno/v3/types"
)

// StoreValidatorFromMsgCreateValidator handles properly a MsgCreateValidator instance by
// saving into the database all the data associated to such validator
func (m *Module) StoreValidatorsFromMsgCreateValidator(height int64, msg *stakingtypes.MsgCreateValidator) error {
	var pubKey cryptotypes.PubKey
	err := m.cdc.UnpackAny(msg.Pubkey, &pubKey)
	if err != nil {
		return fmt.Errorf("error while unpacking pub key: %s", err)
	}
	avatarURL, err := keybase.GetAvatarURL(msg.Description.Identity)
	if err != nil {
		return fmt.Errorf("error while getting Avatar URL: %s", err)
	}

	consAddr, err := utils.ConvertAddressPrefix("likevalcons", sdk.ConsAddress(pubKey.Address()).String())
	if err != nil {
		return fmt.Errorf("error while converting to likevalcons prefix: %s", err)
	}
	operAddr, err := utils.ConvertAddressPrefix("likevaloper", msg.ValidatorAddress)
	if err != nil {
		return fmt.Errorf("error while converting to likevaloper prefix: %s", err)
	}
	selfDelegateAddress, err := utils.ConvertAddressPrefix("like", msg.DelegatorAddress)
	if err != nil {
		return fmt.Errorf("error while converting to like prefix: %s", err)
	}

	// Save the validators
	err = m.db.SaveValidatorData(
		types.NewValidator(
			consAddr,
			operAddr,
			pubKey.String(),
			selfDelegateAddress,
			&msg.Commission.MaxChangeRate,
			&msg.Commission.MaxRate,
			height,
		),
	)
	if err != nil {
		return err
	}

	// Save the descriptions
	err = m.db.SaveValidatorDescription(
		types.NewValidatorDescription(
			operAddr,
			msg.Description,
			avatarURL,
			height,
		),
	)
	if err != nil {
		return err
	}

	// Save the commissions
	err = m.db.SaveValidatorCommission(
		types.NewValidatorCommission(
			operAddr,
			&msg.Commission.Rate,
			&msg.MinSelfDelegation,
			height,
		),
	)
	if err != nil {
		return err
	}

	return nil
}
