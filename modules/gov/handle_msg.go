package gov

import (
	"fmt"

	"strconv"

	"github.com/forbole/bdjuno/v3/modules/utils"
	"github.com/forbole/bdjuno/v3/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	juno "github.com/forbole/juno/v3/types"
)

// HandleMsg implements modules.MessageModule
func (m *Module) HandleMsg(index int, msg sdk.Msg, tx *juno.Tx) error {
	if len(tx.Logs) == 0 {
		return nil
	}

	switch cosmosMsg := msg.(type) {
	case *govtypes.MsgSubmitProposal:
		return m.handleMsgSubmitProposal(tx, index, cosmosMsg)

	case *govtypes.MsgDeposit:
		return m.handleMsgDeposit(tx, cosmosMsg)

	case *govtypes.MsgVote:
		return m.handleMsgVote(tx, cosmosMsg)
	}

	return nil
}

// handleMsgSubmitProposal allows to properly handle a handleMsgSubmitProposal
func (m *Module) handleMsgSubmitProposal(tx *juno.Tx, index int, msg *govtypes.MsgSubmitProposal) error {
	// Get the proposal id
	event, err := tx.FindEventByType(index, govtypes.EventTypeSubmitProposal)
	if err != nil {
		return fmt.Errorf("error while searching for EventTypeSubmitProposal: %s", err)
	}

	id, err := tx.FindAttributeByKey(event, govtypes.AttributeKeyProposalID)
	if err != nil {
		return fmt.Errorf("error while searching for AttributeKeyProposalID: %s", err)
	}

	proposalID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return fmt.Errorf("error while parsing proposal id: %s", err)
	}

	// Get the proposal
	proposal, err := m.source.Proposal(tx.Height, proposalID)
	if err != nil {
		return fmt.Errorf("error while getting proposal: %s", err)
	}

	// Unpack the content
	var content govtypes.Content
	err = m.cdc.UnpackAny(proposal.Content, &content)
	if err != nil {
		return fmt.Errorf("error while unpacking proposal content: %s", err)
	}

	proposerAddress, err := utils.ConvertAddressPrefix("like", msg.Proposer)
	if err != nil {
		return fmt.Errorf("error while converting to like prefix: %s", err)
	}

	// Store the proposal
	proposalObj := types.NewProposal(
		proposal.ProposalId,
		proposal.ProposalRoute(),
		proposal.ProposalType(),
		proposal.GetContent(),
		proposal.Status.String(),
		proposal.SubmitTime,
		proposal.DepositEndTime,
		proposal.VotingStartTime,
		proposal.VotingEndTime,
		proposerAddress,
	)
	err = m.db.SaveProposals([]types.Proposal{proposalObj})
	if err != nil {
		return err
	}

	fmt.Printf("Proposal %d submitted at height %d \n", proposalID, tx.Height)

	// Store the deposit
	deposit := types.NewDeposit(proposal.ProposalId, proposerAddress, msg.InitialDeposit, 1)
	return m.db.SaveDeposits([]types.Deposit{deposit})
}

// handleMsgDeposit allows to properly handle a handleMsgDeposit
func (m *Module) handleMsgDeposit(tx *juno.Tx, msg *govtypes.MsgDeposit) error {

	depositorAddress, err := utils.ConvertAddressPrefix("like", msg.Depositor)
	if err != nil {
		return fmt.Errorf("error while converting to like prefix: %s", err)
	}

	fmt.Printf("proposal deposit to proposal %d at height %d from %s \n", msg.ProposalId, tx.Height, depositorAddress)
	deposits, err := m.source.ProposalDeposits(tx.Height, msg.ProposalId)
	if err != nil {
		return fmt.Errorf("error while getting proposal deposit: %s", err)
	}

	var deposit *govtypes.Deposit
	for _, d := range deposits {
		if d.Depositor == msg.Depositor {
			deposit = &d
			break
		}
	}

	if deposit == nil {
		return fmt.Errorf("error while getting proposal deposit: %s", err)
	}

	return m.db.SaveDeposits([]types.Deposit{
		types.NewDeposit(msg.ProposalId, depositorAddress, deposit.Amount, 1),
	})
}

// handleMsgVote allows to properly handle a handleMsgVote
func (m *Module) handleMsgVote(tx *juno.Tx, msg *govtypes.MsgVote) error {
	voterAddress, err := utils.ConvertAddressPrefix("like", msg.Voter)
	if err != nil {
		return fmt.Errorf("error while converting to like prefix: %s", err)
	}
	fmt.Printf("proposal vote to proposal %d at height %d from %s \n", msg.ProposalId, tx.Height, voterAddress)
	vote := types.NewVote(msg.ProposalId, voterAddress, msg.Option, 1)
	return m.db.SaveVote(vote)
}
