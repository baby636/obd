package bean

import (
	"LightningOnOmni/bean/chainhash"
	"LightningOnOmni/bean/enum"
)

type Message struct {
	Type      int    `json:"type"`
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
	Data      string `json:"data"`
}

//type = 1
type User struct {
	Id       int            `storm:"id,increment" `
	Email    string         `json:"email"`
	Password string         `json:"password"`
	State    enum.UserState `json:"state"`
}

//type = -32
type OpenChannelInfo struct {
	Chain_hash                    chainhash.ChainHash `json:"chain_hash"`
	Temporary_channel_id          chainhash.Hash      `json:"temporary_channel_id"`
	funding_satoshis              uint64              `json:"funding_satoshis"`
	push_msat                     uint64              `json:"push_msat"`
	dust_limit_satoshis           uint64              `json:"dust_limit_satoshis"`
	max_htlc_value_in_flight_msat uint64              `json:"max_htlc_value_in_flight_msat"`
	channel_reserve_satoshis      uint64              `json:"channel_reserve_satoshis"`
	htlc_minimum_msat             uint64              `json:"htlc_minimum_msat"`
	feerate_per_kw                uint32              `json:"feerate_per_kw"`
	to_self_delay                 uint16              `json:"to_self_delay"`
	max_accepted_htlcs            uint16              `json:"max_accepted_htlcs"`
	funding_pubkey                chainhash.Point     `json:"funding_pubkey"`
	revocation_basepoint          chainhash.Point     `json:"revocation_basepoint"`
	payment_basepoint             chainhash.Point     `json:"payment_basepoint"`
	delayed_payment_basepoint     chainhash.Point     `json:"delayed_payment_basepoint"`
	htlc_basepoint                chainhash.Point     `json:"htlc_basepoint"`
}

//type = -33
type AcceptChannelInfo OpenChannelInfo

//type: -38 (close_channel)
type Close_channel struct {
	channel_id   chainhash.Hash
	len          uint16
	scriptpubkey []byte
	signature    chainhash.Signauture
}

//type: -34 (funding_created)
type Funding_created struct {
	Id                   int            `storm:"id,increment" `
	Temporary_channel_id chainhash.Hash `json:"temporaryChannelId"`
	Funder_pubKey        chainhash.Hash `json:"funderPubKey"`
	Property_id          int64          `json:"propertyId"`
	Max_assets           float64        `json:"maxAssets"`
	Amount_a             float64        `json:"amountA"`
}

//type: -35 (funding_signed)
type Funding_signed struct {
	Id int `storm:"id,increment" `
	//the same as the temporary_channel_id in the open_channel message
	Temporary_channel_id chainhash.ChainHash `json:"temporary_channel_id"`
	//the omni address of funder Alice
	Funder_pubKey chainhash.Hash `json:"funder_pub_key"`
	// the id of the Omni asset
	Property_id int `json:"property_id"`
	//amount of the asset on Alice side
	Amount_a float64 `json:"amount_a"`
	//the omni address of fundee Bob
	Fundee_pubKey chainhash.Hash `json:"fundee_pub_key"`
	//amount of the asset on Bob side
	Amount_b float64 `json:"amount_b"`
	//signature of fundee Bob
	Fundee_signature chainhash.Signauture `json:"fundee_signature"`
	//redeem script used to generate P2SH address
	RedeemScript string `json:"redeem_script"`
	//hash of redeemScript
	P2sh_address chainhash.Hash `json:"p_2_sh_address"`
	//final global channel id generated
	Channel_id chainhash.Hash `json:"channel_id"`
}

//type: -351 (commitment_tx)
type Commitment_tx struct {
	Id int `storm:"id,increment" `
	//the global channel id.
	Channel_id chainhash.Hash `json:"channel_id"`
	//the id of the Omni asset
	Property_id int `json:"property_id"`
	//amount of the payment
	Amount float64 `json:"amount"`
	//private key of Alice2, encrypted by Bob's public key
	Encrpted_Alice2_private_key chainhash.Hash `json:"encrpted_alice_2_private_key"`
}

//type: -352 (commitment_tx_signed)
type Commitment_tx_signed struct {
	Id int `storm:"id,increment" `

	//the global channel id.
	Channel_id chainhash.Hash `json:"channel_id"`
	//the id of the Omni asset.
	Property_id int `json:"property_id"`
	//amount of the payment.
	Amount float64 `json:"amount"`
	//signature of Bob.
	Receiver_signature chainhash.Signauture `json:"receiver_signature"`
}

//type: -353 (get_balance_request)
type Get_balance_request struct {
	Id int `storm:"id,increment" `

	//the global channel id.
	Channel_id chainhash.Hash `json:"channel_id"`
	//the p2sh address generated in funding_signed message.
	P2sh_address chainhash.Hash `json:"p_2_sh_address"`
	// the channel owner, Alice or Bob, can query the balance.
	Who chainhash.Hash `json:"who"`
	//the signature of Alice or Bob
	Signature chainhash.Signauture `json:"signature"`
}

//type: -354 (get_balance_respond)
type Get_balance_respond struct {
	Id int `storm:"id,increment" `
	//the global channel id.
	Channel_id chainhash.Hash `json:"channel_id"`
	//the asset id generated by Omnilayer protocol.
	Property_id int `json:"property_id"`
	//the name of the asset.
	Name string `json:"name"`
	//balance in this channel
	Balance float64 `json:"balance"`
	//currently not in use
	Reserved float64 `json:"reserved"`
	//currently not in use
	Frozen float64 `json:"frozen"`
}
