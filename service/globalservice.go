package service

import (
	"LightningOnOmni/bean"
	"LightningOnOmni/dao"
	"LightningOnOmni/rpc"
	"LightningOnOmni/tool"
	"errors"
	"github.com/asdine/storm/q"
	"log"
	"time"

	"github.com/asdine/storm"
	"github.com/tidwall/gjson"
)

var db *storm.DB
var rpcClient *rpc.Client

//for store the privateKey
var tempAddrPrivateKeyMap = make(map[string]string)
var OnlineUserMap = make(map[string]bool)

func FindUserIsOnline(peerId string) error {
	if tool.CheckIsString(&peerId) {
		value, exists := OnlineUserMap[peerId]
		if exists && value == true {
			return nil
		}
	}
	return errors.New("user not exist or online")
}

type commitmentOutputBean struct {
	RsmcTempPubKey   string
	AmountToRsmc     float64
	ToChannelPubKey  string
	ToChannelAddress string
	AmountToOther    float64
	HtlcTempPubKey   string
	AmountToHtlc     float64
}

func init() {
	var err error
	db, err = dao.DBService.GetDB()
	if err != nil {
		log.Println(err)
	}
	rpcClient = rpc.NewClient()
}

func getAddressFromPubKey(pubKey string) (address string, err error) {
	if tool.CheckIsString(&pubKey) == false {
		return "", errors.New("empty pubKey")
	}
	address, err = tool.GetAddressFromPubKey(pubKey)
	if err != nil {
		return "", err
	}
	isValid, err := rpcClient.ValidateAddress(address)
	if err != nil {
		return "", err
	}
	if isValid == false {
		return "", errors.New("invalid pubKey")
	}
	return address, nil
}

func createCommitmentTx(owner string, channelInfo *dao.ChannelInfo, fundingTransaction *dao.FundingTransaction, outputBean commitmentOutputBean, user *bean.User) (*dao.CommitmentTransaction, error) {
	commitmentTxInfo := &dao.CommitmentTransaction{}
	commitmentTxInfo.PeerIdA = channelInfo.PeerIdA
	commitmentTxInfo.PeerIdB = channelInfo.PeerIdB
	commitmentTxInfo.ChannelId = channelInfo.ChannelId
	commitmentTxInfo.PropertyId = fundingTransaction.PropertyId
	commitmentTxInfo.Owner = owner

	//input
	commitmentTxInfo.InputTxid = fundingTransaction.FundingTxid
	commitmentTxInfo.InputVout = fundingTransaction.FundingOutputIndex
	commitmentTxInfo.InputAmount = fundingTransaction.AmountA + fundingTransaction.AmountB

	//output to rsmc
	commitmentTxInfo.RSMCTempAddressPubKey = outputBean.RsmcTempPubKey
	multiAddr, err := rpcClient.CreateMultiSig(2, []string{commitmentTxInfo.RSMCTempAddressPubKey, outputBean.ToChannelPubKey})
	if err != nil {
		return nil, err
	}
	commitmentTxInfo.RSMCMultiAddress = gjson.Get(multiAddr, "address").String()
	commitmentTxInfo.RSMCRedeemScript = gjson.Get(multiAddr, "redeemScript").String()
	json, err := rpcClient.GetAddressInfo(commitmentTxInfo.RSMCMultiAddress)
	if err != nil {
		return nil, err
	}
	commitmentTxInfo.RSMCMultiAddressScriptPubKey = gjson.Get(json, "scriptPubKey").String()

	if tool.CheckIsString(&outputBean.HtlcTempPubKey) {
		commitmentTxInfo.HTLCTempAddressPubKey = outputBean.HtlcTempPubKey
		multiAddr, err := rpcClient.CreateMultiSig(2, []string{commitmentTxInfo.HTLCTempAddressPubKey, outputBean.ToChannelPubKey})
		if err != nil {
			return nil, err
		}
		commitmentTxInfo.HTLCMultiAddress = gjson.Get(multiAddr, "address").String()
		commitmentTxInfo.HTLCRedeemScript = gjson.Get(multiAddr, "redeemScript").String()
		json, err := rpcClient.GetAddressInfo(commitmentTxInfo.HTLCMultiAddress)
		if err != nil {
			return nil, err
		}
		commitmentTxInfo.HTLCMultiAddressScriptPubKey = gjson.Get(json, "scriptPubKey").String()
	}

	commitmentTxInfo.AmountToRSMC = outputBean.AmountToRsmc
	commitmentTxInfo.AmountToOther = outputBean.AmountToOther
	commitmentTxInfo.AmountToHtlc = outputBean.AmountToHtlc

	commitmentTxInfo.CreateBy = user.PeerId
	commitmentTxInfo.CreateAt = time.Now()
	commitmentTxInfo.LastEditTime = time.Now()

	return commitmentTxInfo, nil
}

func createRDTx(owner string, channelInfo *dao.ChannelInfo, commitmentTxInfo *dao.CommitmentTransaction, toAddress string, user *bean.User) (*dao.RevocableDeliveryTransaction, error) {
	rda := &dao.RevocableDeliveryTransaction{}

	rda.CommitmentTxId = commitmentTxInfo.Id
	rda.PeerIdA = channelInfo.PeerIdA
	rda.PeerIdB = channelInfo.PeerIdB
	rda.ChannelId = channelInfo.ChannelId
	rda.PropertyId = commitmentTxInfo.PropertyId
	rda.Owner = owner

	//input
	rda.InputTxid = commitmentTxInfo.RSMCTxid
	rda.InputVout = 0
	rda.InputAmount = commitmentTxInfo.AmountToRSMC
	//output
	rda.OutputAddress = toAddress
	rda.Sequence = 1000
	rda.Amount = commitmentTxInfo.AmountToRSMC

	rda.CreateBy = user.PeerId
	rda.CreateAt = time.Now()
	rda.LastEditTime = time.Now()

	return rda, nil
}
func createBRTx(owner string, channelInfo *dao.ChannelInfo, commitmentTxInfo *dao.CommitmentTransaction, user *bean.User) (*dao.BreachRemedyTransaction, error) {
	breachRemedyTransaction := &dao.BreachRemedyTransaction{}
	breachRemedyTransaction.CommitmentTxId = commitmentTxInfo.Id
	breachRemedyTransaction.PeerIdA = channelInfo.PeerIdA
	breachRemedyTransaction.PeerIdB = channelInfo.PeerIdB
	breachRemedyTransaction.ChannelId = channelInfo.ChannelId
	breachRemedyTransaction.PropertyId = commitmentTxInfo.PropertyId
	breachRemedyTransaction.Owner = owner

	//input
	breachRemedyTransaction.InputTxid = commitmentTxInfo.RSMCTxid
	breachRemedyTransaction.InputVout = 0
	breachRemedyTransaction.InputAmount = commitmentTxInfo.AmountToRSMC
	//output
	breachRemedyTransaction.Amount = commitmentTxInfo.AmountToRSMC

	breachRemedyTransaction.CreateBy = user.PeerId
	breachRemedyTransaction.CreateAt = time.Now()
	breachRemedyTransaction.LastEditTime = time.Now()

	return breachRemedyTransaction, nil
}

func checkBtcTxHex(btcFeeTxHexDecode string, channelInfo *dao.ChannelInfo, peerId string) (fundingTxid string, amountA float64, fundingOutputIndex uint32, err error) {
	jsonFundingTxHexDecode := gjson.Parse(btcFeeTxHexDecode)
	fundingTxid = jsonFundingTxHexDecode.Get("txid").String()

	//vin
	if jsonFundingTxHexDecode.Get("vin").IsArray() == false {
		err = errors.New("wrong Tx input vin")
		log.Println(err)
		return "", 0, 0, err
	}
	inTxid := jsonFundingTxHexDecode.Get("vin").Array()[0].Get("txid").String()
	inputTx, err := rpcClient.GetTransactionById(inTxid)
	if err != nil {
		err = errors.New("wrong input: " + err.Error())
		log.Println(err)
		return "", 0, 0, err
	}

	jsonInputTxDecode := gjson.Parse(inputTx)
	flag := false
	inputHexDecode, err := rpcClient.DecodeRawTransaction(jsonInputTxDecode.Get("hex").String())
	if err != nil {
		err = errors.New("wrong input: " + err.Error())
		log.Println(err)
		return "", 0, 0, err
	}

	funderAddress := channelInfo.AddressA
	if peerId == channelInfo.PeerIdB {
		funderAddress = channelInfo.AddressB
	}
	jsonInputHexDecode := gjson.Parse(inputHexDecode)
	if jsonInputHexDecode.Get("vout").IsArray() {
		for _, item := range jsonInputHexDecode.Get("vout").Array() {
			addresses := item.Get("scriptPubKey").Get("addresses").Array()
			for _, subItem := range addresses {
				if subItem.String() == funderAddress {
					flag = true
					break
				}
			}
			if flag {
				break
			}
		}
	}

	if flag == false {
		err = errors.New("wrong vin " + jsonFundingTxHexDecode.Get("vin").String())
		log.Println(err)
		return "", 0, 0, err
	}

	//vout
	flag = false
	if jsonFundingTxHexDecode.Get("vout").IsArray() == false {
		err = errors.New("wrong Tx vout")
		log.Println(err)
		return "", 0, 0, err
	}
	for _, item := range jsonFundingTxHexDecode.Get("vout").Array() {
		addresses := item.Get("scriptPubKey").Get("addresses").Array()
		for _, subItem := range addresses {
			if subItem.String() == channelInfo.ChannelAddress {
				amountA = item.Get("value").Float()
				fundingOutputIndex = uint32(item.Get("n").Int())
				flag = true
				break
			}
		}
		if flag {
			break
		}
	}
	if flag == false {
		err = errors.New("wrong vout " + jsonFundingTxHexDecode.Get("vout").String())
		log.Println(err)
		return "", 0, 0, err
	}
	return fundingTxid, amountA, fundingOutputIndex, err
}

func checkOmniTxHex(fundingTxHexDecode string, channelInfo *dao.ChannelInfo, user *bean.User) (fundingTxid string, amountA float64, propertyId int64, err error) {
	jsonOmniTxHexDecode := gjson.Parse(fundingTxHexDecode)
	fundingTxid = jsonOmniTxHexDecode.Get("txid").String()

	funderAddress := channelInfo.AddressA
	if user.PeerId == channelInfo.PeerIdB {
		funderAddress = channelInfo.AddressB
	}

	sendingAddress := jsonOmniTxHexDecode.Get("sendingaddress").String()
	if sendingAddress != funderAddress {
		err = errors.New("wrong Tx input")
		log.Println(err)
		return "", 0, 0, err
	}
	referenceAddress := jsonOmniTxHexDecode.Get("referenceaddress").String()
	if referenceAddress != channelInfo.ChannelAddress {
		err = errors.New("wrong Tx output")
		log.Println(err)
		return "", 0, 0, err
	}

	amountA = jsonOmniTxHexDecode.Get("amount").Float()
	propertyId = jsonOmniTxHexDecode.Get("propertyid").Int()

	return fundingTxid, amountA, propertyId, err
}

//从未广播的交易hash数据中解析出他的输出，以此作为下个交易的输入
func getInputsOfNextTxByParseTxHashVout(hex string, toAddress, scriptPubKey string) (inputs []rpc.TransactionInputItem, err error) {
	result, err := rpcClient.DecodeRawTransaction(hex)
	if err != nil {
		return nil, err
	}
	jsonHex := gjson.Parse(result)
	log.Println(jsonHex)
	if jsonHex.Get("vout").IsArray() {
		inputs = make([]rpc.TransactionInputItem, 0, 0)
		for _, item := range jsonHex.Get("vout").Array() {
			if item.Get("scriptPubKey").Get("addresses").Exists() {
				addresses := item.Get("scriptPubKey").Get("addresses").Array()
				for _, address := range addresses {
					if address.String() == toAddress {
						node := rpc.TransactionInputItem{}
						node.Txid = jsonHex.Get("txid").String()
						node.ScriptPubKey = scriptPubKey
						node.Vout = uint32(item.Get("n").Uint())
						node.Amount = item.Get("value").Float()
						inputs = append(inputs, node)
					}
				}
			}
		}
		return inputs, nil
	}
	return nil, errors.New("no inputs")
}

func getLatestCommitmentTx(channelId bean.ChannelID, owner string) (commitmentTxInfo *dao.CommitmentTransaction, err error) {
	commitmentTxInfo = &dao.CommitmentTransaction{}
	err = db.Select(q.Eq("ChannelId", channelId), q.Eq("Owner", owner), q.Eq("CurrState", dao.TxInfoState_CreateAndSign)).OrderBy("CreateAt").Reverse().First(commitmentTxInfo)
	return commitmentTxInfo, err
}
